// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "forge-std/console.sol";
import "../src/AppleToken.sol";
import "../src/AppleAMM.sol";

/**
 * @title Deploy Script
 * @notice Deploys AppleToken and AppleAMM, then seeds with initial liquidity.
 * @dev Invoked by scripts/deploy.sh (and Ubuntu reseed via Go handlers).
 *      Broadcast JSON is read by backend/internal/config for contract addresses.
 *
 * Account allocation (Anvil accounts 0-29):
 * 0: Deployer/LP - Seeds pool with 25,000 APPL + 25,000 ETH (initial price: 1 ETH/APPL)
 *     Gets 30,000 ETH total (25k liquidity + 5k for gas)
 * 1: User - Manual trading account (5,000 ETH and 5,000 APPL)
 * 2-16: Retail - 15 retail trading bots (1,000 ETH and 1,000 APPL each)
 * 17-19: Whale - 3 whale trading bots (1,000 ETH and 5,000 APPL each)
 * 20-22: MeanRev - 3 mean reversion bots (1,000 ETH and 1,000 APPL each)
 * 23-29: Reserved - Available for future use (1,000 ETH and 1,000 APPL each)
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
        console.log("\n=== Deploying Contracts ===");
        console.log("Note: 'Estimated total gas used' shows gas units (not ETH/gwei)");
        console.log("      'Estimated amount required' shows total ETH needed for gas fees");
        console.log("      (Forge displays this with many decimals - typically ~0.01 ETH for deployment)");
        console.log("");
        token = new AppleToken();
        console.log("AppleToken deployed at:", address(token));
        amm = new AppleAMM(address(token));
        console.log("AppleAMM deployed at:", address(amm));
        console.log("Note: Transaction hashes for deployments appear in the transaction log below");
        
        // Redistribute ETH balances FIRST (before minting to save gas)
        // Account 0 needs extra ETH (10k for liquidity + gas)
        // This must happen inside broadcast so it's actually set on-chain
        redistributeETH(accounts);
        
        // Mint tokens to all accounts
        mintTokens(accounts);
        
        // Log account balances table
        logAccountBalances(accounts);
        
        // Seed initial liquidity (account 0 = LP)
        // This happens in the same broadcast block so account 0 has the redistributed ETH
        seedLiquidity();
        
        vm.stopBroadcast();
        
        // Log final state
        logState();
    }
    
    function getAnvilAccounts() internal pure returns (address[] memory) {
        address[] memory accounts = new address[](30);
        
        // These are the default Anvil addresses, reorganized to match new account layout:
        // 0: LP, 1: User, 2-16: Retail, 17-19: Whale, 20-22: MeanRev, 23-29: Reserved
        
        accounts[0] = 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266;  // LP (index 0)
        accounts[1] = 0x70997970C51812dc3A010C7d01b50e0d17dc79C8;  // User (index 1)
        
        // Retail accounts (2-16) - use addresses from indices 2-16
        accounts[2] = 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC;   // Retail1 (index 2)
        accounts[3] = 0x90F79bf6EB2c4f870365E785982E1f101E93b906;   // Retail2 (index 3)
        accounts[4] = 0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65;   // Retail3 (index 4)
        accounts[5] = 0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc;   // Retail4 (index 5)
        accounts[6] = 0x976EA74026E726554dB657fA54763abd0C3a0aa9;   // Retail5 (index 6)
        accounts[7] = 0x14dC79964da2C08b23698B3D3cc7Ca32193d9955;   // Retail6 (index 7)
        accounts[8] = 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f;   // Retail7 (index 8)
        accounts[9] = 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720;   // Retail8 (index 9)
        accounts[10] = 0xBcd4042DE499D14e55001CcbB24a551F3b954096;  // Retail9 (index 10)
        accounts[11] = 0x71bE63f3384f5fb98995898A86B02Fb2426c5788;  // Retail10 (index 11)
        accounts[12] = 0xFABB0ac9d68B0B445fB7357272Ff202C5651694a;  // Retail11 (index 12)
        accounts[13] = 0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec;  // Retail12 (index 13)
        accounts[14] = 0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097;  // Retail13 (index 14)
        accounts[15] = 0xcd3B766CCDd6AE721141F452C550Ca635964ce71;  // Retail14 (index 15)
        accounts[16] = 0x2546BcD3c84621e976D8185a91A922aE77ECEc30;  // Retail15 (index 16)
        
        // Whale accounts (17-19) - use addresses from indices 17-19
        accounts[17] = 0xbDA5747bFD65F08deb54cb465eB87D40e51B197E;  // Whale1 (index 17)
        accounts[18] = 0xdD2FD4581271e230360230F9337D5c0430Bf44C0;  // Whale2 (index 18)
        accounts[19] = 0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199;  // Whale3 (index 19)
        
        // MeanRev accounts (20-22) - use addresses from indices 20-22
        accounts[20] = 0x09DB0a93B389bEF724429898f539AEB7ac2Dd55f;  // MeanRev1 (index 20)
        accounts[21] = 0x02484cb50AAC86Eae85610D6f4Bf026f30f6627D;  // MeanRev2 (index 21)
        accounts[22] = 0x08135Da0A343E492FA2d4282F2AE34c6c5CC1BbE;  // MeanRev3 (index 22)
        
        // Reserved accounts (23-29) - use addresses from indices 23-29
        accounts[23] = 0x5E661B79FE2D3F6cE70F5AAC07d8Cd9abb2743F1;  // Reserved (index 23)
        accounts[24] = 0x61097BA76cD906d2ba4FD106E757f7Eb455fc295;  // Reserved (index 24)
        accounts[25] = 0xDf37F81dAAD2b0327A0A50003740e1C935C70913;  // Reserved (index 25)
        accounts[26] = 0x553BC17A05702530097c3677091C5BB47a3a7931;  // Reserved (index 26)
        accounts[27] = 0x87BdCE72c06C21cd96219BD8521bDF1F42C78b5e;  // Reserved (index 27)
        accounts[28] = 0x40Fc963A729c542424cD800349a7E4Ecc4896624;  // Reserved (index 28)
        accounts[29] = 0x9DCCe783B6464611f38631e6C851bf441907c710;  // Reserved (index 29)
        
        return accounts;
    }
    
    function mintTokens(address[] memory accounts) internal {
        console.log("\n=== Minting Tokens ===");
        
        // LP gets tokens for seeding pool
        token.mint(accounts[0], 25000 * APPLES_DECIMALS);
        console.log("Minting to LP Account 0: 25,000 APPL");
        
        // User account
        token.mint(accounts[1], 5000 * APPLES_DECIMALS);
        console.log("Minting to User Account 1: 5,000 APPL");
        
        // Retail accounts (2-16)
        for (uint i = 2; i <= 16; i++) {
            token.mint(accounts[i], 1000 * APPLES_DECIMALS);
        }
        console.log("Minting to Retail accounts (2-16): 1,000 APPL each");
        
        // Whale accounts (17-19)
        for (uint i = 17; i <= 19; i++) {
            token.mint(accounts[i], 5000 * APPLES_DECIMALS);
        }
        console.log("Minting to Whale accounts (17-19): 5,000 APPL each");
        
        // MeanRev accounts (20-22)
        for (uint i = 20; i <= 22; i++) {
            token.mint(accounts[i], 1000 * APPLES_DECIMALS);
        }
        console.log("Minting to MeanRev accounts (20-22): 1,000 APPL each");
        
        // Reserved accounts (23-29)
        for (uint i = 23; i <= 29; i++) {
            token.mint(accounts[i], 1000 * APPLES_DECIMALS);
        }
        console.log("Minting to Reserved accounts (23-29): 1,000 APPL each");
        
        console.log("\nTokens minted to all accounts");
    }
    
    function redistributeETH(address[] memory accounts) internal {
        console.log("\n=== Setting ETH Balances ===");
        console.log("Note: 'Block:' shows the block number where the transaction was included");
        console.log("      'Paid:' shows gas cost = (gas used * gas price in gwei) / 1e9 ETH");
        console.log("      Gas price varies by network conditions (Anvil uses low prices for testing)\n");
        
        // Use vm.deal to set balances for simulation
        // Account 0 (LP) needs 30k ETH total (25k for liquidity + 5k for gas)
        vm.deal(accounts[0], 30000 ether);
        console.log("Setting ETH balance for LP Account 0: 30,000 ETH");
        
        // User account gets 5k ETH for larger manual trade experiments
        vm.deal(accounts[1], 5000 ether);
        console.log("Setting ETH balance for User Account 1: 5,000 ETH");

        // All remaining accounts get 1k ETH
        for (uint i = 2; i < accounts.length; i++) {
            vm.deal(accounts[i], 1000 ether);
        }
        console.log("Setting ETH balance for accounts 2-29: 1,000 ETH each");
        
        console.log("\nETH balances set for simulation:");
        console.log("  Account 0 (LP) = 30,000 ETH");
        console.log("  Account 1 (User) = 5,000 ETH");
        console.log("  Accounts 2-29 = 1,000 ETH each");
        console.log("Note: For actual deployment, ensure account 0 has sufficient ETH for liquidity + gas");
    }
    
    function logAccountBalances(address[] memory accounts) internal view {
        console.log("\n=== Account Starting Balances ===");
        console.log("Account          | ETH Balance | APPL Balance");
        console.log("-----------------|-------------|-------------");
        
        for (uint i = 0; i < accounts.length; i++) {
            uint256 ethBalance = accounts[i].balance;
            uint256 applBalance = token.balanceOf(accounts[i]);
            uint256 ethWhole = ethBalance / 1e18;
            uint256 applWhole = applBalance / APPLES_DECIMALS;
            
            // Log each account with balances
            if (i == 0) {
                console.log("LP 0 - ETH:", ethWhole);
                console.log("LP 0 - APPL:", applWhole);
            } else if (i == 1) {
                console.log("User 1 - ETH:", ethWhole);
                console.log("User 1 - APPL:", applWhole);
            } else if (i >= 2 && i <= 16) {
                uint256 retailNum = i - 1;
                console.log("Retail", retailNum, "- ETH:", ethWhole);
                console.log("Retail", retailNum, "- APPL:", applWhole);
            } else if (i >= 17 && i <= 19) {
                uint256 whaleNum = i - 16;
                console.log("Whale", whaleNum, "- ETH:", ethWhole);
                console.log("Whale", whaleNum, "- APPL:", applWhole);
            } else if (i >= 20 && i <= 22) {
                uint256 meanRevNum = i - 19;
                console.log("MeanRev", meanRevNum, "- ETH:", ethWhole);
                console.log("MeanRev", meanRevNum, "- APPL:", applWhole);
            } else {
                console.log("Reserved", i, "- ETH:", ethWhole);
                console.log("Reserved", i, "- APPL:", applWhole);
            }
        }
    }
    
    function getAccountNickname(uint256 index) internal pure returns (string memory) {
        if (index == 0) return "LP";
        if (index == 1) return "User";
        if (index >= 2 && index <= 16) {
            uint256 retailNum = index - 1;
            return string.concat("Retail", vm.toString(retailNum));
        }
        if (index >= 17 && index <= 19) {
            uint256 whaleNum = index - 16;
            return string.concat("Whale", vm.toString(whaleNum));
        }
        if (index >= 20 && index <= 22) {
            uint256 meanRevNum = index - 19;
            return string.concat("MeanRev", vm.toString(meanRevNum));
        }
        return "Reserved";
    }
    
    function seedLiquidity() internal {
        // LP seeds pool with 25,000 APPL + 25,000 ETH
        // Initial price = 1 ETH/APPL (1:1 ratio)
        // Large pool size needed for whale trades (500-600 APPL clips)
        token.approve(address(amm), 25000 * APPLES_DECIMALS);
        amm.addLiquidity{value: 25000 ether}(25000 * APPLES_DECIMALS);

        console.log("Pool seeded with 25,000 APPL + 25,000 ETH (initial price: 1 ETH/APPL)");
    }
    
    function logState() internal view {
        (uint256 appleRes, uint256 ethRes) = amm.getReserves();
        uint256 price = amm.getSpotPrice();
        uint256 lpTokens = amm.totalLPTokens();
        
        console.log("\n=== Deployment Complete ===");
        console.log("Contract Addresses:");
        console.log("  AppleToken:", address(token));
        console.log("  AppleAMM:", address(amm));
        console.log("");
        console.log("Note: The verbose transaction log above (with hashes, blocks, gas costs) is Forge's");
        console.log("      --broadcast output showing all transactions sent to the network.");
        console.log("      This is expected and cannot be suppressed without removing --broadcast.");
        console.log("      Look for 'Hash:' entries to find individual transaction IDs.");
        console.log("");
        console.log("Pool State:");
        console.log("  Apple Reserve:", appleRes / APPLES_DECIMALS, "APPL");
        console.log("  ETH Reserve:", ethRes / 1e18, "ETH");
        console.log("  Spot Price:", price / 1e18, "ETH per APPL");
        console.log("  LP Tokens:", lpTokens / 1e18, "LP (in ETH terms)");
        console.log("  LP Tokens (raw):", lpTokens);
    }
}
