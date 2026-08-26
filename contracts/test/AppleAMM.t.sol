// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/AppleToken.sol";
import "../src/AppleAMM.sol";

/**
 * @title AppleAMM Tests
 * @notice Core AMM behavior: liquidity add/remove, buys/sells, fee, invariant.
 * @dev Mirrors the constant-product path used by Go bots (engine/executor.go)
 *      and the frontend impact curve. Run with `make test-contracts` / forge test.
 */
contract AppleAMMTest is Test {
    AppleToken public token;
    AppleAMM public amm;
    
    address public lp;
    address public trader;
    
    uint256 constant INITIAL_APPLES = 1000 ether;
    uint256 constant INITIAL_ETH = 1000 ether;
    
    function setUp() public {
        lp = makeAddr("lp");
        trader = makeAddr("trader");
        
        // Deploy contracts
        token = new AppleToken();
        amm = new AppleAMM(address(token));
        
        // Setup LP
        vm.deal(lp, 10000 ether);
        token.mint(lp, 10000 ether);
        
        // Setup trader
        vm.deal(trader, 1000 ether);
        token.mint(trader, 1000 ether);
        
        // Initial liquidity from LP
        vm.startPrank(lp);
        token.approve(address(amm), type(uint256).max);
        amm.addLiquidity{value: INITIAL_ETH}(INITIAL_APPLES);
        vm.stopPrank();
        
        // Trader approves AMM
        vm.prank(trader);
        token.approve(address(amm), type(uint256).max);
    }
    
    // ============================================================
    // LIQUIDITY TESTS
    // ============================================================
    
    function test_InitialLiquidity() public view {
        (uint256 appleRes, uint256 ethRes) = amm.getReserves();
        
        assertEq(appleRes, INITIAL_APPLES);
        assertEq(ethRes, INITIAL_ETH);
        assertGt(amm.totalLPTokens(), 0);
        assertEq(amm.lpBalances(lp), amm.totalLPTokens());
    }
    
    function test_SpotPrice() public view {
        // Initial ratio is 1:1, so price should be 1e18
        uint256 price = amm.getSpotPrice();
        assertEq(price, 1e18);
    }
    
    function test_AddMoreLiquidity() public {
        uint256 lpTokensBefore = amm.lpBalances(lp);
        
        vm.prank(lp);
        amm.addLiquidity{value: 100 ether}(100 ether);
        
        (uint256 appleRes, uint256 ethRes) = amm.getReserves();
        assertEq(appleRes, 1100 ether);
        assertEq(ethRes, 1100 ether);
        assertGt(amm.lpBalances(lp), lpTokensBefore);
    }
    
    function test_RemoveLiquidity() public {
        uint256 lpTokens = amm.lpBalances(lp);
        uint256 halfTokens = lpTokens / 2;
        
        uint256 lpAppleBefore = token.balanceOf(lp);
        uint256 lpEthBefore = lp.balance;
        
        vm.prank(lp);
        (uint256 appleOut, uint256 ethOut) = amm.removeLiquidity(halfTokens);
        
        // Should get back ~half the reserves
        assertApproxEqRel(appleOut, INITIAL_APPLES / 2, 0.01e18);
        assertApproxEqRel(ethOut, INITIAL_ETH / 2, 0.01e18);
        
        // Check balances updated
        assertEq(token.balanceOf(lp), lpAppleBefore + appleOut);
        assertEq(lp.balance, lpEthBefore + ethOut);
    }
    
    // ============================================================
    // SWAP TESTS
    // ============================================================
    
    function test_SwapETHForApples() public {
        uint256 ethIn = 10 ether;
        
        uint256 traderAppleBefore = token.balanceOf(trader);
        
        vm.prank(trader);
        uint256 applesOut = amm.swapETHForApples{value: ethIn}(0);
        
        // Trader should have received apples
        assertGt(applesOut, 0);
        assertEq(token.balanceOf(trader), traderAppleBefore + applesOut);
        
        // Reserves should have changed
        (uint256 appleRes, uint256 ethRes) = amm.getReserves();
        assertLt(appleRes, INITIAL_APPLES);
        assertGt(ethRes, INITIAL_ETH);
    }
    
    function test_SwapApplesForETH() public {
        uint256 applesIn = 10 ether;
        
        uint256 traderEthBefore = trader.balance;
        
        vm.prank(trader);
        uint256 ethOut = amm.swapApplesForETH(applesIn, 0);
        
        // Trader should have received ETH
        assertGt(ethOut, 0);
        assertEq(trader.balance, traderEthBefore + ethOut);
        
        // Reserves should have changed
        (uint256 appleRes, uint256 ethRes) = amm.getReserves();
        assertGt(appleRes, INITIAL_APPLES);
        assertLt(ethRes, INITIAL_ETH);
    }
    
    function test_SwapPricing() public {
        // For constant product, output should follow x*y=k
        uint256 ethIn = 100 ether;
        
        // Calculate expected output
        uint256 fee = (ethIn * 30) / 10000;  // 0.30%
        uint256 ethInWithFee = ethIn - fee;
        uint256 expectedOut = amm.getAmountOut(ethInWithFee, INITIAL_ETH, INITIAL_APPLES);
        
        vm.prank(trader);
        uint256 actualOut = amm.swapETHForApples{value: ethIn}(0);
        
        assertEq(actualOut, expectedOut);
    }
    
    function test_SlippageProtection() public {
        uint256 ethIn = 10 ether;
        uint256 minApples = 100 ether;  // Unrealistic minimum
        
        vm.prank(trader);
        vm.expectRevert(AppleAMM.SlippageExceeded.selector);
        amm.swapETHForApples{value: ethIn}(minApples);
    }
    
    // ============================================================
    // INVARIANT TESTS
    // ============================================================
    
    function test_InvariantPreserved() public {
        // Get initial k
        (uint256 apple0, uint256 eth0) = amm.getReserves();
        uint256 k0 = apple0 * eth0;
        
        // Do a swap
        vm.prank(trader);
        amm.swapETHForApples{value: 50 ether}(0);
        
        // k should increase (due to fees staying in pool)
        (uint256 apple1, uint256 eth1) = amm.getReserves();
        uint256 k1 = apple1 * eth1;
        
        assertGe(k1, k0, "k should not decrease after swap");
    }
    
    function test_MultipleSwapsInvariant() public {
        (uint256 apple0, uint256 eth0) = amm.getReserves();
        uint256 k0 = apple0 * eth0;
        
        // Buy apples
        vm.prank(trader);
        amm.swapETHForApples{value: 20 ether}(0);
        
        // Sell apples
        vm.prank(trader);
        amm.swapApplesForETH(30 ether, 0);
        
        // Buy more
        vm.prank(trader);
        amm.swapETHForApples{value: 10 ether}(0);
        
        (uint256 apple1, uint256 eth1) = amm.getReserves();
        uint256 k1 = apple1 * eth1;
        
        // k should only grow due to fees
        assertGe(k1, k0);
    }
    
    // ============================================================
    // FEE TESTS
    // ============================================================
    
    function test_FeeAccumulation() public {
        (uint256 feeApple0, uint256 feeETH0) = amm.getTotalFees();
        assertEq(feeApple0, 0);
        assertEq(feeETH0, 0);
        
        // Buy (pays ETH fee)
        vm.prank(trader);
        amm.swapETHForApples{value: 100 ether}(0);
        
        (uint256 feeApple1, uint256 feeETH1) = amm.getTotalFees();
        assertEq(feeApple1, 0);
        assertGt(feeETH1, 0);
        assertEq(feeETH1, (100 ether * 30) / 10000);  // 0.30%
        
        // Sell (pays APPL fee)
        vm.prank(trader);
        amm.swapApplesForETH(50 ether, 0);
        
        (uint256 feeApple2, uint256 feeETH2) = amm.getTotalFees();
        assertGt(feeApple2, 0);
        assertEq(feeETH2, feeETH1);  // ETH fees unchanged
        assertEq(feeApple2, (50 ether * 30) / 10000);  // 0.30%
    }
    
    function test_FeesRemainInPool() public {
        // Record initial reserves
        (uint256 apple0, uint256 eth0) = amm.getReserves();
        
        // Do swaps
        vm.prank(trader);
        amm.swapETHForApples{value: 100 ether}(0);
        
        vm.prank(trader);
        amm.swapApplesForETH(100 ether, 0);
        
        // Reserves should be larger than initial due to fees
        (uint256 apple1, uint256 eth1) = amm.getReserves();
        assertGe(apple1 + eth1, apple0 + eth0);
    }
    
    // ============================================================
    // PRICE IMPACT TEST
    // ============================================================
    
    function test_PriceImpact() public {
        uint256 price0 = amm.getSpotPrice();
        
        // Large buy should increase price
        vm.prank(trader);
        amm.swapETHForApples{value: 200 ether}(0);
        
        uint256 price1 = amm.getSpotPrice();
        assertGt(price1, price0, "Price should increase after large buy");
        
        // Large sell should decrease price
        vm.prank(trader);
        amm.swapApplesForETH(300 ether, 0);
        
        uint256 price2 = amm.getSpotPrice();
        assertLt(price2, price1, "Price should decrease after large sell");
    }
    
    // ============================================================
    // EDGE CASES
    // ============================================================
    
    function test_RevertOnEmptyPool() public {
        // Deploy fresh AMM without liquidity
        AppleAMM emptyAmm = new AppleAMM(address(token));
        
        vm.prank(trader);
        vm.expectRevert(AppleAMM.PoolEmpty.selector);
        emptyAmm.swapETHForApples{value: 1 ether}(0);
    }
    
    function test_LargeTradeHighSlippage() public {
        // Very large trades are possible but have extreme slippage
        // This is expected behavior of constant product AMM
        vm.deal(trader, 100000 ether);
        
        uint256 priceBefore = amm.getSpotPrice();
        
        vm.prank(trader);
        uint256 applesOut = amm.swapETHForApples{value: 50000 ether}(0);
        
        // You can buy most of the pool but not all
        assertLt(applesOut, INITIAL_APPLES, "Can't drain entire pool");
        
        uint256 priceAfter = amm.getSpotPrice();
        
        // Price should increase massively
        assertGt(priceAfter, priceBefore * 10, "Price should spike after huge buy");
    }
}
