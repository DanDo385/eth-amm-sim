// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title AppleAMM
 * @notice Constant product AMM (x * y = k) for APPL/ETH pair
 * @dev 0.30% fee (30 bps) on all swaps, fee remains in pool
 * 
 * Key mechanics:
 * - Liquidity providers deposit both APPL and ETH in proportion to current reserves
 * - LP tokens track pro-rata share of the pool
 * - Swaps maintain the x*y=k invariant (minus fees)
 * - Fees accumulate in the pool, benefiting all LPs
 */
contract AppleAMM is ReentrancyGuard {
    using SafeERC20 for IERC20;
    
    // ============================================================
    // STATE
    // ============================================================
    
    IERC20 public immutable appleToken;
    
    // Pool reserves
    uint256 public appleReserve;
    uint256 public ethReserve;
    
    // LP token accounting (simplified, no ERC20)
    uint256 public totalLPTokens;
    mapping(address => uint256) public lpBalances;
    
    // Fee: 0.30% = 30 bps = 30/10000
    uint256 public constant FEE_NUMERATOR = 30;
    uint256 public constant FEE_DENOMINATOR = 10000;
    
    // Track fees for metrics
    uint256 public totalFeesApple;
    uint256 public totalFeesETH;
    
    // ============================================================
    // EVENTS
    // ============================================================
    
    event LogSwap(
        address indexed trader,
        bool isBuy,              // true = buying APPL with ETH
        uint256 amountIn,
        uint256 amountOut,
        uint256 price,           // price after swap (scaled by 1e18)
        uint256 fee,
        uint256 timestamp
    );
    
    event LogLiquidityAdded(
        address indexed provider,
        uint256 appleAmount,
        uint256 ethAmount,
        uint256 lpTokens
    );
    
    event LogLiquidityRemoved(
        address indexed provider,
        uint256 appleAmount,
        uint256 ethAmount,
        uint256 lpTokens
    );
    
    // ============================================================
    // CONSTRUCTOR
    // ============================================================
    
    constructor(address _appleToken) {
        require(_appleToken != address(0), "AppleAMM: invalid token address");
        appleToken = IERC20(_appleToken);
    }
    
    // ============================================================
    // LIQUIDITY FUNCTIONS
    // ============================================================
    
    /**
     * @notice Add liquidity to the pool
     * @param appleAmount Amount of APPL to deposit
     * @return lpTokens Amount of LP tokens minted
     * 
     * @dev First deposit sets the initial ratio.
     *      Subsequent deposits must maintain the current ratio.
     *      ETH is sent as msg.value.
     */
    function addLiquidity(uint256 appleAmount) external payable nonReentrant returns (uint256 lpTokens) {
        require(appleAmount > 0 && msg.value > 0, "AppleAMM: amounts must be positive");
        
        // Transfer APPL from sender
        appleToken.safeTransferFrom(msg.sender, address(this), appleAmount);
        
        if (totalLPTokens == 0) {
            // First deposit: LP tokens = sqrt(appleAmount * ethAmount)
            // Using geometric mean to prevent manipulation
            lpTokens = sqrt(appleAmount * msg.value);
            require(lpTokens > 0, "AppleAMM: insufficient initial liquidity");
        } else {
            // Subsequent deposits: maintain ratio
            // LP tokens proportional to contribution
            uint256 lpFromApple = (appleAmount * totalLPTokens) / appleReserve;
            uint256 lpFromETH = (msg.value * totalLPTokens) / ethReserve;
            
            // Use minimum to prevent gaming
            lpTokens = lpFromApple < lpFromETH ? lpFromApple : lpFromETH;
            require(lpTokens > 0, "AppleAMM: insufficient liquidity");
        }
        
        // Update state
        appleReserve += appleAmount;
        ethReserve += msg.value;
        totalLPTokens += lpTokens;
        lpBalances[msg.sender] += lpTokens;
        
        emit LogLiquidityAdded(msg.sender, appleAmount, msg.value, lpTokens);
    }
    
    /**
     * @notice Remove liquidity from the pool
     * @param lpTokenAmount Amount of LP tokens to burn
     * @return appleOut Amount of APPL returned
     * @return ethOut Amount of ETH returned
     */
    function removeLiquidity(uint256 lpTokenAmount) external nonReentrant returns (uint256 appleOut, uint256 ethOut) {
        require(lpTokenAmount > 0, "AppleAMM: amount must be positive");
        require(lpBalances[msg.sender] >= lpTokenAmount, "AppleAMM: insufficient LP tokens");
        
        // Calculate pro-rata share
        appleOut = (lpTokenAmount * appleReserve) / totalLPTokens;
        ethOut = (lpTokenAmount * ethReserve) / totalLPTokens;
        
        require(appleOut > 0 && ethOut > 0, "AppleAMM: insufficient output");
        
        // Update state BEFORE transfers (reentrancy protection)
        lpBalances[msg.sender] -= lpTokenAmount;
        totalLPTokens -= lpTokenAmount;
        appleReserve -= appleOut;
        ethReserve -= ethOut;
        
        // Transfer assets
        appleToken.safeTransfer(msg.sender, appleOut);
        (bool success, ) = msg.sender.call{value: ethOut}("");
        require(success, "AppleAMM: ETH transfer failed");
        
        emit LogLiquidityRemoved(msg.sender, appleOut, ethOut, lpTokenAmount);
    }
    
    // ============================================================
    // SWAP FUNCTIONS
    // ============================================================
    
    /**
     * @notice Swap ETH for APPL (buy APPL)
     * @param minApples Minimum APPL to receive (slippage protection)
     * @return applesOut Amount of APPL received
     */
    function swapETHForApples(uint256 minApples) external payable nonReentrant returns (uint256 applesOut) {
        require(msg.value > 0, "AppleAMM: ETH amount must be positive");
        require(appleReserve > 0 && ethReserve > 0, "AppleAMM: pool is empty");
        
        // Calculate output with fee
        uint256 fee = (msg.value * FEE_NUMERATOR) / FEE_DENOMINATOR;
        uint256 amountInWithFee = msg.value - fee;
        
        applesOut = getAmountOut(amountInWithFee, ethReserve, appleReserve);
        require(applesOut >= minApples, "AppleAMM: slippage exceeded");
        require(applesOut < appleReserve, "AppleAMM: insufficient liquidity");
        
        // Update reserves
        ethReserve += msg.value;  // Full amount including fee
        appleReserve -= applesOut;
        totalFeesETH += fee;
        
        // Transfer APPL to buyer
        appleToken.safeTransfer(msg.sender, applesOut);
        
        emit LogSwap(
            msg.sender,
            true,  // isBuy
            msg.value,
            applesOut,
            getSpotPrice(),
            fee,
            block.timestamp
        );
    }
    
    /**
     * @notice Swap APPL for ETH (sell APPL)
     * @param appleAmount Amount of APPL to sell
     * @param minETH Minimum ETH to receive (slippage protection)
     * @return ethOut Amount of ETH received
     */
    function swapApplesForETH(uint256 appleAmount, uint256 minETH) external nonReentrant returns (uint256 ethOut) {
        require(appleAmount > 0, "AppleAMM: APPL amount must be positive");
        require(appleReserve > 0 && ethReserve > 0, "AppleAMM: pool is empty");
        
        // Transfer APPL from seller
        appleToken.safeTransferFrom(msg.sender, address(this), appleAmount);
        
        // Calculate output with fee
        uint256 fee = (appleAmount * FEE_NUMERATOR) / FEE_DENOMINATOR;
        uint256 amountInWithFee = appleAmount - fee;
        
        ethOut = getAmountOut(amountInWithFee, appleReserve, ethReserve);
        require(ethOut >= minETH, "AppleAMM: slippage exceeded");
        require(ethOut < ethReserve, "AppleAMM: insufficient liquidity");
        
        // Update reserves
        appleReserve += appleAmount;  // Full amount including fee
        ethReserve -= ethOut;
        totalFeesApple += fee;
        
        // Transfer ETH to seller
        (bool success, ) = msg.sender.call{value: ethOut}("");
        require(success, "AppleAMM: ETH transfer failed");
        
        emit LogSwap(
            msg.sender,
            false,  // isBuy (false = sell)
            appleAmount,
            ethOut,
            getSpotPrice(),
            fee,
            block.timestamp
        );
    }
    
    // ============================================================
    // VIEW FUNCTIONS
    // ============================================================
    
    /**
     * @notice Get current pool reserves
     * @return _appleReserve Current APPL reserve
     * @return _ethReserve Current ETH reserve
     */
    function getReserves() external view returns (uint256 _appleReserve, uint256 _ethReserve) {
        return (appleReserve, ethReserve);
    }
    
    /**
     * @notice Get current spot price (ETH per APPL, scaled by 1e18)
     * @return price Price in wei of ETH per 1 APPL
     * @dev Returns ethReserve * 1e18 / appleReserve
     */
    function getSpotPrice() public view returns (uint256 price) {
        if (appleReserve == 0) return 0;
        return (ethReserve * 1e18) / appleReserve;
    }
    
    /**
     * @notice Calculate output amount for a given input (constant product formula)
     * @param amountIn Input amount (after fee deduction)
     * @param reserveIn Reserve of input token
     * @param reserveOut Reserve of output token
     * @return amountOut Output amount
     * @dev Uses x * y = k formula
     *      amountOut = (amountIn * reserveOut) / (reserveIn + amountIn)
     */
    function getAmountOut(
        uint256 amountIn,
        uint256 reserveIn,
        uint256 reserveOut
    ) public pure returns (uint256 amountOut) {
        require(amountIn > 0, "AppleAMM: insufficient input amount");
        require(reserveIn > 0 && reserveOut > 0, "AppleAMM: insufficient liquidity");
        
        // Constant product formula: x * y = k
        // newReserveIn = reserveIn + amountIn
        // newReserveOut = k / newReserveIn = (reserveIn * reserveOut) / (reserveIn + amountIn)
        // amountOut = reserveOut - newReserveOut = (amountIn * reserveOut) / (reserveIn + amountIn)
        
        uint256 numerator = amountIn * reserveOut;
        uint256 denominator = reserveIn + amountIn;
        amountOut = numerator / denominator;
    }
    
    /**
     * @notice Get LP token balance for an address
     * @param account Address to query
     * @return balance LP token balance
     */
    function getLPBalance(address account) external view returns (uint256 balance) {
        return lpBalances[account];
    }
    
    /**
     * @notice Get total fees collected
     * @return feesApple Total APPL fees
     * @return feesETH Total ETH fees
     */
    function getTotalFees() external view returns (uint256 feesApple, uint256 feesETH) {
        return (totalFeesApple, totalFeesETH);
    }
    
    // ============================================================
    // INTERNAL FUNCTIONS
    // ============================================================
    
    /**
     * @notice Calculate square root using Babylonian method
     * @param x Value to take square root of
     * @return y Square root of x
     */
    function sqrt(uint256 x) internal pure returns (uint256 y) {
        if (x == 0) return 0;
        
        uint256 z = (x + 1) / 2;
        y = x;
        while (z < y) {
            y = z;
            z = (x / z + z) / 2;
        }
    }
    
    // Allow contract to receive ETH
    receive() external payable {}
}
