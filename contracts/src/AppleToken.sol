// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title AppleToken
 * @notice Simple ERC20 token for AMM simulation
 * @dev Owner can mint tokens for initial distribution
 */
contract AppleToken is ERC20, Ownable {
    constructor() ERC20("Apple", "APPL") Ownable(msg.sender) {}
    
    /**
     * @notice Mint tokens to an address
     * @param to Recipient address
     * @param amount Amount to mint (in wei)
     */
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
}
