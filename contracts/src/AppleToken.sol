// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title AppleToken
 * @notice Simple ERC20 token for AMM simulation
 * @dev Owner can mint tokens for initial distribution
 */
contract AppleToken is ERC20 {
    address public owner;
    
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    modifier onlyOwner() {
        require(msg.sender == owner, "AppleToken: caller is not the owner");
        _;
    }
    
    constructor() ERC20("Apple", "APPL") {
        owner = msg.sender;
    }
    
    /**
     * @notice Mint tokens to an address
     * @param to Recipient address
     * @param amount Amount to mint (in wei)
     */
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
    
    /**
     * @notice Transfer ownership to a new address
     * @param newOwner New owner address
     */
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "AppleToken: new owner is the zero address");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }
}
