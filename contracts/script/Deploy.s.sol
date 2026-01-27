// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/AppleToken.sol";
import "../src/AppleAMM.sol";

/**
 * @title Deploy Script
 * @notice Deploys AppleToken and AppleAMM, then seeds with initial liquidity
 * 
 * Account allocation (Anvil accounts 0-29):
 * 0: Deployer/LP - Seeds pool with 10,000 APPL + 10,000 ETH (initial price: 1 ETH/APPL)
 *     Gets 15,000 ETH total (10k liquidity + 5k for loans/gas)
 * 1-29: All other accounts get 1,000 ETH and 1,000 APPL each
 *     Exception: Account 28 (Liquidator) gets 0 APPL (uses ETH only)
 *     Leverage accounts (25-27) use their ETH as collateral for leveraged positions
 */
contract DeployScript is Script {
    // Anvil default private keys (DO NOT use in production!)
    // Account 0: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
    
    AppleToken public token;
    AppleAMM public amm;
    
    uint256 constant APPLES_DECIMALS = 1e18;
    
    function run() external {
        // Get deployer private key (account 0)
        uint256 deployerKey = vm.envOr(
            "PRIVATE_KEY",
            uint256(0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80)
        );
        
        // Get Anvil account addresses
        address[] memory accounts = getAnvilAccounts();
        
        vm.startBroadcast(deployerKey);
        
        // Deploy contracts
        token = new AppleToken();
        amm = new AppleAMM(address(token));
        
        console.log("AppleToken deployed at:", address(token));
        console.log("AppleAMM deployed at:", address(amm));
        
        // Mint tokens to all accounts
        mintTokens(accounts);
        
        vm.stopBroadcast();
        
        // Redistribute ETH balances (using separate broadcast blocks)
        // Account 0 needs extra ETH (10k for liquidity + gas)
        // Leverage accounts (25-27) only need 1k ETH each (not shorting)
        redistributeETH(accounts);
        
        vm.startBroadcast(deployerKey);
        
        // Seed initial liquidity (account 0 = LP)
        seedLiquidity();
        
        vm.stopBroadcast();
        
        // Log final state
        logState();
    }
    
    function getAnvilAccounts() internal pure returns (address[] memory) {
        address[] memory accounts = new address[](30);
        
        // These are the default Anvil addresses
        accounts[0] = 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266;  // LP
        accounts[1] = 0x70997970C51812dc3A010C7d01b50e0d17dc79C8;  // Whale1
        accounts[2] = 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC;  // Whale2
        accounts[3] = 0x90F79bf6EB2c4f870365E785982E1f101E93b906;  // Whale3
        accounts[4] = 0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65;  // MeanRev1
        accounts[5] = 0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc;  // MeanRev2
        accounts[6] = 0x976EA74026E726554dB657fA54763abd0C3a0aa9;  // MeanRev3
        accounts[7] = 0x14dC79964da2C08b23698B3D3cc7Ca32193d9955;  // Unused (reserved)
        accounts[8] = 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f;  // Unused (reserved)
        accounts[9] = 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720;  // Unused (reserved)
        
        // Retail accounts (10-24)
        accounts[10] = 0xBcd4042DE499D14e55001CcbB24a551F3b954096;
        accounts[11] = 0x71bE63f3384f5fb98995898A86B02Fb2426c5788;
        accounts[12] = 0xFABB0ac9d68B0B445fB7357272Ff202C5651694a;
        accounts[13] = 0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec;
        accounts[14] = 0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097;
        accounts[15] = 0xcd3B766CCDd6AE721141F452C550Ca635964ce71;
        accounts[16] = 0x2546BcD3c84621e976D8185a91A922aE77ECEc30;
        accounts[17] = 0xbDA5747bFD65F08deb54cb465eB87D40e51B197E;
        accounts[18] = 0xdD2FD4581271e230360230F9337D5c0430Bf44C0;
        accounts[19] = 0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199;
        accounts[20] = 0x09DB0a93B389bEF724429898f539AEB7ac2Dd55f;
        accounts[21] = 0x02484cb50AAC86Eae85610D6f4Bf026f30f6627D;
        accounts[22] = 0x08135Da0A343E492FA2d4282F2AE34c6c5CC1BbE;
        accounts[23] = 0x5E661B79FE2D3F6cE70F5AAC07d8Cd9abb2743F1;
        accounts[24] = 0x61097BA76cD906d2ba4FD106E757f7Eb455fc295;
        
        // Leveraged traders (25-27)
        accounts[25] = 0xDf37F81dAAD2b0327A0A50003740e1C935C70913;
        accounts[26] = 0x553BC17A05702530097c3677091C5BB47a3a7931;
        accounts[27] = 0x87BdCE72c06C21cd96219BD8521bDF1F42C78b5e;
        
        // Liquidator (28)
        accounts[28] = 0x40Fc963A729c542424cD800349a7E4Ecc4896624;
        
        // Reserved (29)
        accounts[29] = 0x9DCCe783B6464611f38631e6C851bf441907c710;
        
        return accounts;
    }
    
    function mintTokens(address[] memory accounts) internal {
        // LP gets tokens for seeding pool
        token.mint(accounts[0], 10000 * APPLES_DECIMALS);
        
        // All other accounts get 1,000 APPL each (except liquidator)
        // Whales
        token.mint(accounts[1], 1000 * APPLES_DECIMALS);  // Whale1
        token.mint(accounts[2], 1000 * APPLES_DECIMALS);  // Whale2
        token.mint(accounts[3], 1000 * APPLES_DECIMALS);  // Whale3
        
        // Mean reversion bots
        token.mint(accounts[4], 1000 * APPLES_DECIMALS);  // MeanRev1
        token.mint(accounts[5], 1000 * APPLES_DECIMALS); // MeanRev2
        token.mint(accounts[6], 1000 * APPLES_DECIMALS); // MeanRev3
        
        // Accounts 7-9: Unused (reserved for future use)
        for (uint i = 7; i <= 9; i++) {
            token.mint(accounts[i], 1000 * APPLES_DECIMALS);
        }
        
        // Retail traders (10-24)
        for (uint i = 10; i <= 24; i++) {
            token.mint(accounts[i], 1000 * APPLES_DECIMALS);
        }
        
        // Leveraged traders (25-27) - get 1k APPL each (they use ETH for collateral)
        token.mint(accounts[25], 1000 * APPLES_DECIMALS); // Lev5x
        token.mint(accounts[26], 1000 * APPLES_DECIMALS); // Lev10x
        token.mint(accounts[27], 1000 * APPLES_DECIMALS); // Lev25x
        
        // Liquidator (28) gets 0 APPL (use ETH)
        
        // User account (29)
        token.mint(accounts[29], 1000 * APPLES_DECIMALS);
        
        console.log("Tokens minted to all accounts");
    }
    
    function redistributeETH(address[] memory accounts) internal {
        // Use vm.deal to set balances for simulation
        // Account 0 (LP) needs 15k ETH total (10k for liquidity + 5k for loans/gas)
        vm.deal(accounts[0], 15000 ether);
        
        // All other accounts get 1k ETH (leverage accounts included - they only need collateral)
        // Note: Leverage accounts will use their ETH as collateral for leveraged positions
        for (uint i = 1; i < accounts.length; i++) {
            vm.deal(accounts[i], 1000 ether);
        }
        
        console.log("ETH balances set for simulation: Account 0 (LP) = 15k ETH, All other accounts = 1k ETH each");
        console.log("Note: For actual deployment, ensure account 0 has sufficient ETH for liquidity + gas");
    }
    
    function seedLiquidity() internal {
        // LP seeds pool with 10,000 APPL + 10,000 ETH
        // Initial price = 1 ETH/APPL (1:1 ratio)
        // Large pool size needed for whale trades (500-600 APPL clips)
        token.approve(address(amm), 10000 * APPLES_DECIMALS);
        amm.addLiquidity{value: 10000 ether}(10000 * APPLES_DECIMALS);
        
        console.log("Pool seeded with 10,000 APPL + 10,000 ETH (initial price: 1 ETH/APPL)");
    }
    
    function logState() internal view {
        (uint256 appleRes, uint256 ethRes) = amm.getReserves();
        uint256 price = amm.getSpotPrice();
        
        console.log("\n=== Deployment Complete ===");
        console.log("Apple Reserve:", appleRes / APPLES_DECIMALS, "APPL");
        console.log("ETH Reserve:", ethRes / 1e18, "ETH");
        console.log("Spot Price:", price / 1e18, "ETH per APPL");
        console.log("LP Tokens:", amm.totalLPTokens());
    }
}
