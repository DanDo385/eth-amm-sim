// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Address.sol";

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
    // CUSTOM ERRORS
    // ============================================================
    
    error InsufficientLiquidity();
    error InsufficientAPPLLiquidity();
    error InsufficientETHLiquidity();
    error InvalidLeverage();
    error PositionNotLiquidatable();
    error PositionAlreadyExists();
    error NoPosition();
    error PoolEmpty();
    error InvalidTokenAddress();
    error AmountMustBePositive();
    error InsufficientLPTokens();
    error InsufficientOutput();
    error SlippageExceeded();
    error InsufficientInitialMargin();
    error OnlyLongPositionsSupported();
    error InsufficientInputAmount();
    
    // ============================================================
    // CONSTANTS
    // ============================================================
    
    uint256 private constant PRICE_SCALE = 1e18;
    uint256 private constant MAX_LEVERAGE = 25;
    uint256 private constant MIN_LEVERAGE = 1;
    
    // Fee: 0.30% = 30 bps = 30/10000
    uint256 public constant FEE_NUMERATOR = 30;
    uint256 public constant FEE_DENOMINATOR = 10000;
    
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
    
    // Track fees for metrics
    uint256 public totalFeesApple;
    uint256 public totalFeesETH;
    
    // ============================================================
    // LEVERAGED POSITIONS
    // ============================================================
    
    struct Position {
        address trader;
        bool isLong;              // true = long APPL, false = short APPL (not implemented)
        uint256 collateral;       // ETH collateral deposited
        uint256 size;             // Position size in APPL (for long positions)
        uint256 entryPrice;       // Entry price (ETH per APPL, scaled by 1e18)
        uint256 leverage;         // Leverage multiplier (5x, 10x, 25x)
        uint256 timestamp;        // Position opened timestamp
    }
    
    mapping(address => Position) public positions;
    
    // Margin requirements (in basis points, 10000 = 100%)
    uint256 public constant INITIAL_MARGIN_BPS = 5000;      // 50% initial margin
    uint256 public constant MAINTENANCE_MARGIN_BPS = 2000;   // 20% maintenance margin
    uint256 public constant LIQUIDATION_THRESHOLD_BPS = 1000; // 10% liquidation threshold
    uint256 public constant LIQUIDATION_FEE_BPS = 500;       // 5% liquidation fee
    uint256 private constant BASIS_POINTS = 10000;
    
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
    
    event PositionOpened(
        address indexed trader,
        bool isLong,
        uint256 collateral,
        uint256 size,
        uint256 entryPrice,
        uint256 leverage
    );
    
    event PositionClosed(
        address indexed trader,
        uint256 collateralReturned,
        uint256 pnl
    );
    
    event PositionLiquidated(
        address indexed trader,
        address indexed liquidator,
        uint256 collateral,
        uint256 liquidationFee,
        uint256 remainingCollateral
    );
    
    // ============================================================
    // CONSTRUCTOR
    // ============================================================
    
    constructor(address _appleToken) {
        if (_appleToken == address(0)) revert InvalidTokenAddress();
        appleToken = IERC20(_appleToken);
    }
    
    // ============================================================
    // LIQUIDITY FUNCTIONS
    // ============================================================
    
    /**
     * @notice Add liquidity to the pool
     * @param appleAmount Amount of APPL tokens to deposit
     * @return lpTokens Amount of LP tokens minted to the provider
     * 
     * @dev First deposit sets the initial ratio using geometric mean (sqrt(x*y)).
     *      Subsequent deposits must maintain the current reserve ratio.
     *      ETH is sent as msg.value and must match the APPL/ETH ratio.
     *      LP tokens represent pro-rata share of the pool.
     */
    function addLiquidity(uint256 appleAmount) external payable nonReentrant returns (uint256 lpTokens) {
        if (appleAmount == 0 || msg.value == 0) revert AmountMustBePositive();
        
        // Transfer APPL from sender
        appleToken.safeTransferFrom(msg.sender, address(this), appleAmount);
        
        if (totalLPTokens == 0) {
            // First deposit: LP tokens = sqrt(appleAmount * ethAmount)
            // Using geometric mean to prevent manipulation
            lpTokens = sqrt(appleAmount * msg.value);
            if (lpTokens == 0) revert InsufficientLiquidity();
        } else {
            // Subsequent deposits: maintain ratio
            // LP tokens proportional to contribution
            uint256 lpFromApple = (appleAmount * totalLPTokens) / appleReserve;
            uint256 lpFromETH = (msg.value * totalLPTokens) / ethReserve;
            
            // Use minimum to prevent gaming
            lpTokens = lpFromApple < lpFromETH ? lpFromApple : lpFromETH;
            if (lpTokens == 0) revert InsufficientLiquidity();
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
     * @return appleOut Amount of APPL tokens returned to provider
     * @return ethOut Amount of ETH returned to provider
     * 
     * @dev Returns pro-rata share of both reserves based on LP token amount.
     *      State is updated before transfers to prevent reentrancy.
     */
    function removeLiquidity(uint256 lpTokenAmount) external nonReentrant returns (uint256 appleOut, uint256 ethOut) {
        if (lpTokenAmount == 0) revert AmountMustBePositive();
        if (lpBalances[msg.sender] < lpTokenAmount) revert InsufficientLPTokens();
        
        // Calculate pro-rata share
        appleOut = (lpTokenAmount * appleReserve) / totalLPTokens;
        ethOut = (lpTokenAmount * ethReserve) / totalLPTokens;
        
        if (appleOut == 0 || ethOut == 0) revert InsufficientOutput();
        
        // Update state BEFORE transfers (reentrancy protection)
        lpBalances[msg.sender] -= lpTokenAmount;
        totalLPTokens -= lpTokenAmount;
        appleReserve -= appleOut;
        ethReserve -= ethOut;
        
        // Transfer assets
        appleToken.safeTransfer(msg.sender, appleOut);
        Address.sendValue(payable(msg.sender), ethOut);
        
        emit LogLiquidityRemoved(msg.sender, appleOut, ethOut, lpTokenAmount);
    }
    
    // ============================================================
    // SWAP FUNCTIONS
    // ============================================================
    
    /**
     * @notice Swap ETH for APPL (buy APPL)
     * @param minApples Minimum APPL tokens to receive (slippage protection)
     * @return applesOut Amount of APPL tokens received
     * 
     * @dev ETH is sent as msg.value. A 0.30% fee is deducted before calculating output.
     *      Reverts if output is less than minApples or pool lacks liquidity.
     */
    function swapETHForApples(uint256 minApples) external payable nonReentrant returns (uint256 applesOut) {
        if (msg.value == 0) revert AmountMustBePositive();
        _validatePoolNotEmpty();
        
        // Calculate fee and amount after fee
        (uint256 fee, uint256 amountInWithFee) = _calculateSwapFee(msg.value);
        
        // Calculate output amount
        applesOut = getAmountOut(amountInWithFee, ethReserve, appleReserve);
        if (applesOut < minApples) revert SlippageExceeded();
        if (applesOut >= appleReserve) revert InsufficientAPPLLiquidity();
        
        // Execute swap
        _executeSwap(true, msg.value, applesOut, fee);
        
        return applesOut;
    }
    
    /**
     * @notice Swap APPL for ETH (sell APPL)
     * @param appleAmount Amount of APPL tokens to sell
     * @param minETH Minimum ETH to receive (slippage protection)
     * @return ethOut Amount of ETH received
     * 
     * @dev APPL tokens are transferred from caller. A 0.30% fee is deducted before calculating output.
     *      Reverts if output is less than minETH or pool lacks liquidity.
     */
    function swapApplesForETH(uint256 appleAmount, uint256 minETH) external nonReentrant returns (uint256 ethOut) {
        if (appleAmount == 0) revert AmountMustBePositive();
        _validatePoolNotEmpty();
        
        // Transfer APPL from seller
        appleToken.safeTransferFrom(msg.sender, address(this), appleAmount);
        
        // Calculate fee and amount after fee
        (uint256 fee, uint256 amountInWithFee) = _calculateSwapFee(appleAmount);
        
        // Calculate output amount
        ethOut = getAmountOut(amountInWithFee, appleReserve, ethReserve);
        if (ethOut < minETH) revert SlippageExceeded();
        if (ethOut >= ethReserve) revert InsufficientETHLiquidity();
        
        // Execute swap
        _executeSwap(false, appleAmount, ethOut, fee);
        
        return ethOut;
    }
    
    // ============================================================
    // VIEW FUNCTIONS
    // ============================================================
    
    /**
     * @notice Get current pool reserves
     * @return _appleReserve Current APPL token reserve in the pool
     * @return _ethReserve Current ETH reserve in the pool
     */
    function getReserves() external view returns (uint256 _appleReserve, uint256 _ethReserve) {
        return (appleReserve, ethReserve);
    }
    
    /**
     * @notice Get current spot price (ETH per APPL, scaled by PRICE_SCALE)
     * @return price Price in wei of ETH per 1 APPL
     * @dev Returns ethReserve * PRICE_SCALE / appleReserve
     */
    function getSpotPrice() public view returns (uint256 price) {
        if (appleReserve == 0) return 0;
        return (ethReserve * PRICE_SCALE) / appleReserve;
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
        if (amountIn == 0) revert InsufficientInputAmount();
        if (reserveIn == 0 || reserveOut == 0) revert InsufficientLiquidity();
        
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
     * @param account Address to query LP token balance for
     * @return balance LP token balance of the account
     */
    function getLPBalance(address account) external view returns (uint256 balance) {
        return lpBalances[account];
    }
    
    /**
     * @notice Get total fees collected since contract deployment
     * @return feesApple Total APPL fees collected (from sell swaps)
     * @return feesETH Total ETH fees collected (from buy swaps)
     */
    function getTotalFees() external view returns (uint256 feesApple, uint256 feesETH) {
        return (totalFeesApple, totalFeesETH);
    }
    
    // ============================================================
    // LEVERAGED POSITION FUNCTIONS
    // ============================================================
    
    /**
     * @notice Open a leveraged long position (buy APPL with leverage)
     * @param leverage Leverage multiplier (must be between 1 and 25)
     * @param minApples Minimum APPL tokens to receive (slippage protection)
     * @return applesOut Amount of APPL tokens received
     * 
     * @dev Deposits ETH collateral (msg.value) and executes swap to buy APPL.
     *      Requires 50% initial margin. Leverage affects margin requirements only.
     *      Position is stored for liquidation monitoring.
     */
    function openLeveragedPosition(
        uint256 leverage,
        uint256 minApples
    ) external payable nonReentrant returns (uint256 applesOut) {
        // Validate inputs
        _validateLeverageInputs(leverage, msg.value);
        
        // Execute leveraged swap
        applesOut = _executeLeveragedSwap(msg.value, minApples);
        
        // Create position
        _createPosition(msg.value, applesOut, leverage);
        
        return applesOut;
    }
    
    /**
     * @notice Close a leveraged position
     * @return ethOut Amount of ETH returned to trader
     * 
     * @dev Sells APPL position back to the pool and returns ETH proceeds.
     *      PnL is calculated as (current_value - leveraged_cost).
     *      Only long positions are currently supported.
     */
    function closePosition() external nonReentrant returns (uint256 ethOut) {
        Position memory pos = positions[msg.sender];
        if (pos.collateral == 0) revert NoPosition();
        if (!pos.isLong) revert OnlyLongPositionsSupported();
        
        // Transfer APPL from trader
        appleToken.safeTransferFrom(msg.sender, address(this), pos.size);
        
        // Execute swap to close position
        (uint256 fee, uint256 amountInWithFee) = _calculateSwapFee(pos.size);
        ethOut = getAmountOut(amountInWithFee, appleReserve, ethReserve);
        if (ethOut >= ethReserve) revert InsufficientETHLiquidity();
        
        // Update reserves
        appleReserve += pos.size;
        ethReserve -= ethOut;
        totalFeesApple += fee;
        
        // Calculate PnL
        uint256 leveragedCost = pos.collateral * pos.leverage;
        int256 pnl = int256(ethOut) - int256(leveragedCost);
        
        // Return ETH to trader
        Address.sendValue(payable(msg.sender), ethOut);
        
        // Clear position
        delete positions[msg.sender];
        
        emit PositionClosed(msg.sender, ethOut, uint256(pnl));
        
        // Also emit swap event
        emit LogSwap(
            msg.sender,
            false,
            pos.size,
            ethOut,
            getSpotPrice(),
            fee,
            block.timestamp
        );
        
        return ethOut;
    }
    
    /**
     * @notice Check if a position is liquidatable
     * @param trader Address of the trader to check
     * @return liquidatable True if position can be liquidated (margin ratio < 10%)
     * @return marginRatio Current margin ratio in basis points (10000 = 100%)
     * 
     * @dev Margin ratio = (current_value / leveraged_cost) * 10000
     *      Position is liquidatable if margin ratio < LIQUIDATION_THRESHOLD_BPS (10%)
     */
    function isLiquidatable(address trader) public view returns (bool liquidatable, uint256 marginRatio) {
        Position memory pos = positions[trader];
        if (pos.collateral == 0) {
            return (false, 0);
        }
        
        // Calculate current position value (uses current spot price internally)
        uint256 currentValue = getPositionValue(trader);
        
        // Calculate margin ratio
        marginRatio = _calculateMarginRatio(currentValue, pos.collateral, pos.leverage);
        liquidatable = marginRatio < LIQUIDATION_THRESHOLD_BPS;
        
        return (liquidatable, marginRatio);
    }
    
    /**
     * @notice Get current position value (mark-to-market)
     * @param trader Address of the trader
     * @return value Current value of position in ETH (what would be received if sold now)
     * 
     * @dev Calculates value using current spot price: APPL_amount * current_price
     *      This is the mark-to-market value for liquidation checks.
     */
    function getPositionValue(address trader) public view returns (uint256) {
        Position memory pos = positions[trader];
        if (pos.collateral == 0) {
            return 0;
        }
        
        if (pos.isLong) {
            uint256 currentPrice = getSpotPrice();
            // Long: value = APPL_amount * current_price
            uint256 applValue = (pos.size * currentPrice) / PRICE_SCALE;
            // Return the APPL value (what we'd get if we sold)
            return applValue;
        }
        
        return 0; // Short not implemented
    }
    
    /**
     * @notice Get current position value with explicit price (for backward compatibility)
     * @param trader Address of the trader
     * @param currentPrice Current spot price (ETH per APPL, scaled by PRICE_SCALE)
     * @return value Current value of position in ETH
     * 
     * @dev Allows specifying price explicitly instead of using current spot price.
     *      Useful for testing or when price is already known.
     */
    function getPositionValue(address trader, uint256 currentPrice) public view returns (uint256) {
        Position memory pos = positions[trader];
        if (pos.collateral == 0) {
            return 0;
        }
        
        if (pos.isLong) {
            // Long: value = APPL_amount * current_price
            uint256 applValue = (pos.size * currentPrice) / PRICE_SCALE;
            // Return the APPL value (what we'd get if we sold)
            return applValue;
        }
        
        return 0; // Short not implemented
    }
    
    /**
     * @notice Liquidate an underwater position (anyone can call)
     * @param trader Address of the trader to liquidate
     * @return success True if liquidation was successful
     * 
     * @dev Permissionless liquidation. Sells APPL position, pays 5% liquidation fee to liquidator,
     *      and returns remaining collateral to trader (if any).
     *      Position must be liquidatable (margin ratio < 10%).
     */
    function liquidate(address trader) external nonReentrant returns (bool success) {
        (bool liquidatable, ) = isLiquidatable(trader);
        if (!liquidatable) revert PositionNotLiquidatable();
        
        Position memory pos = positions[trader];
        if (pos.collateral == 0) revert NoPosition();
        
        // Execute liquidation swap
        uint256 ethOut = _executeLiquidationSwap(trader, pos);
        
        // Distribute liquidation proceeds
        _distributeLiquidationProceeds(trader, ethOut, pos.collateral);
        
        // Clear position
        delete positions[trader];
        
        return true;
    }
    
    /**
     * @notice Get position details for a trader
     * @param trader Address of the trader to query
     * @return pos Position struct containing collateral, size, entry price, leverage, etc.
     */
    function getPosition(address trader) external view returns (Position memory pos) {
        return positions[trader];
    }
    
    // ============================================================
    // INTERNAL HELPER FUNCTIONS
    // ============================================================
    
    /**
     * @notice Calculate swap fee and amount after fee
     * @param amountIn Input amount
     * @return fee Fee amount
     * @return amountInWithFee Amount after fee deduction
     */
    function _calculateSwapFee(uint256 amountIn) internal pure returns (uint256 fee, uint256 amountInWithFee) {
        fee = (amountIn * FEE_NUMERATOR) / FEE_DENOMINATOR;
        amountInWithFee = amountIn - fee;
    }
    
    /**
     * @notice Execute swap: update reserves, transfer tokens, emit event
     * @param isBuy True if buying APPL with ETH, false if selling APPL for ETH
     * @param amountIn Input amount (full amount including fee)
     * @param amountOut Output amount
     * @param fee Fee amount
     */
    function _executeSwap(bool isBuy, uint256 amountIn, uint256 amountOut, uint256 fee) internal {
        // Update reserves
        if (isBuy) {
            ethReserve += amountIn;  // Full amount including fee
            appleReserve -= amountOut;
            totalFeesETH += fee;
            // Transfer APPL to buyer
            appleToken.safeTransfer(msg.sender, amountOut);
        } else {
            appleReserve += amountIn;  // Full amount including fee
            ethReserve -= amountOut;
            totalFeesApple += fee;
            // Transfer ETH to seller
            Address.sendValue(payable(msg.sender), amountOut);
        }
        
        // Emit swap event
        emit LogSwap(
            msg.sender,
            isBuy,
            amountIn,
            amountOut,
            getSpotPrice(),
            fee,
            block.timestamp
        );
    }
    
    /**
     * @notice Validate that pool has liquidity
     */
    function _validatePoolNotEmpty() internal view {
        if (appleReserve == 0 || ethReserve == 0) revert PoolEmpty();
    }
    
    // ============================================================
    // LEVERAGED POSITION HELPERS
    // ============================================================
    
    /**
     * @notice Validate leverage inputs
     * @param leverage Leverage multiplier
     * @param collateral Collateral amount
     */
    function _validateLeverageInputs(uint256 leverage, uint256 collateral) internal view {
        if (collateral == 0) revert AmountMustBePositive();
        if (leverage < MIN_LEVERAGE || leverage > MAX_LEVERAGE) revert InvalidLeverage();
        if (positions[msg.sender].collateral > 0) revert PositionAlreadyExists();
        
        // Check initial margin requirement (50% of collateral)
        uint256 requiredMargin = (collateral * INITIAL_MARGIN_BPS) / BASIS_POINTS;
        if (collateral < requiredMargin) revert InsufficientInitialMargin();
    }
    
    /**
     * @notice Execute leveraged swap (buy APPL with collateral)
     * @param tradeAmount ETH amount to trade
     * @param minApples Minimum APPL to receive
     * @return applesOut Amount of APPL received
     */
    function _executeLeveragedSwap(uint256 tradeAmount, uint256 minApples) internal returns (uint256 applesOut) {
        _validatePoolNotEmpty();
        
        // Calculate fee and amount after fee
        (uint256 fee, uint256 amountInWithFee) = _calculateSwapFee(tradeAmount);
        
        // Calculate output amount
        applesOut = getAmountOut(amountInWithFee, ethReserve, appleReserve);
        if (applesOut < minApples) revert SlippageExceeded();
        if (applesOut >= appleReserve) revert InsufficientAPPLLiquidity();
        
        // Update reserves
        ethReserve += tradeAmount;
        appleReserve -= applesOut;
        totalFeesETH += fee;
        
        // Transfer APPL to trader
        appleToken.safeTransfer(msg.sender, applesOut);
        
        // Emit swap event
        uint256 entryPrice = getSpotPrice();
        emit LogSwap(
            msg.sender,
            true,
            tradeAmount,
            applesOut,
            entryPrice,
            fee,
            block.timestamp
        );
        
        return applesOut;
    }
    
    /**
     * @notice Create and store leveraged position
     * @param collateral Collateral amount
     * @param applesOut APPL received
     * @param leverage Leverage multiplier
     */
    function _createPosition(uint256 collateral, uint256 applesOut, uint256 leverage) internal {
        uint256 entryPrice = getSpotPrice();
        positions[msg.sender] = Position({
            trader: msg.sender,
            isLong: true,
            collateral: collateral,
            size: applesOut,
            entryPrice: entryPrice,
            leverage: leverage,
            timestamp: block.timestamp
        });
        
        emit PositionOpened(msg.sender, true, collateral, applesOut, entryPrice, leverage);
    }
    
    /**
     * @notice Execute liquidation swap (sell APPL for ETH)
     * @param trader Address of trader being liquidated
     * @param pos Position to liquidate
     * @return ethOut ETH received from swap
     */
    function _executeLiquidationSwap(address trader, Position memory pos) internal returns (uint256 ethOut) {
        // Transfer APPL from trader
        appleToken.safeTransferFrom(trader, address(this), pos.size);
        
        // Calculate output with fee
        (uint256 fee, uint256 amountInWithFee) = _calculateSwapFee(pos.size);
        ethOut = getAmountOut(amountInWithFee, appleReserve, ethReserve);
        if (ethOut >= ethReserve) revert InsufficientETHLiquidity();
        
        // Update reserves
        appleReserve += pos.size;
        ethReserve -= ethOut;
        totalFeesApple += fee;
        
        // Emit swap event
        emit LogSwap(
            trader,
            false,
            pos.size,
            ethOut,
            getSpotPrice(),
            fee,
            block.timestamp
        );
        
        return ethOut;
    }
    
    /**
     * @notice Distribute liquidation proceeds (fee to liquidator, remainder to trader)
     * @param trader Address of trader being liquidated
     * @param ethOut ETH received from liquidation swap
     * @param collateral Original collateral amount
     */
    function _distributeLiquidationProceeds(address trader, uint256 ethOut, uint256 collateral) internal {
        // Calculate liquidation fee (5% of collateral)
        uint256 liquidationFee = _calculateLiquidationFee(collateral);
        uint256 remainingCollateral = ethOut > liquidationFee ? ethOut - liquidationFee : 0;
        
        // Pay liquidator fee
        Address.sendValue(payable(msg.sender), liquidationFee);
        
        // Return remaining collateral to trader (if any)
        if (remainingCollateral > 0) {
            Address.sendValue(payable(trader), remainingCollateral);
        }
        
        emit PositionLiquidated(trader, msg.sender, collateral, liquidationFee, remainingCollateral);
    }
    
    /**
     * @notice Calculate liquidation fee
     * @param collateral Collateral amount
     * @return fee Liquidation fee amount
     */
    function _calculateLiquidationFee(uint256 collateral) internal pure returns (uint256) {
        return (collateral * LIQUIDATION_FEE_BPS) / BASIS_POINTS;
    }
    
    /**
     * @notice Calculate margin ratio
     * @param currentValue Current position value
     * @param collateral Collateral amount
     * @param leverage Leverage multiplier
     * @return marginRatio Margin ratio in basis points
     */
    function _calculateMarginRatio(uint256 currentValue, uint256 collateral, uint256 leverage) internal pure returns (uint256) {
        uint256 leveragedCost = collateral * leverage;
        if (leveragedCost == 0) {
            return 0;
        }
        return (currentValue * BASIS_POINTS) / leveragedCost;
    }
    
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
