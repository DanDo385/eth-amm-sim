// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/AppleToken.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title AppleToken Tests
 * @notice Basic tests for the AppleToken ERC20 contract
 */
contract AppleTokenTest is Test {
    AppleToken public token;
    address public owner;
    address public alice;
    address public bob;
    
    function setUp() public {
        owner = address(this);
        alice = makeAddr("alice");
        bob = makeAddr("bob");
        
        token = new AppleToken();
    }
    
    function test_InitialState() public view {
        assertEq(token.name(), "Apple");
        assertEq(token.symbol(), "APPL");
        assertEq(token.decimals(), 18);
        assertEq(token.owner(), owner);
        assertEq(token.totalSupply(), 0);
    }
    
    function test_Mint() public {
        uint256 amount = 1000 ether;
        
        token.mint(alice, amount);
        
        assertEq(token.balanceOf(alice), amount);
        assertEq(token.totalSupply(), amount);
    }
    
    function test_MintMultiple() public {
        token.mint(alice, 100 ether);
        token.mint(bob, 200 ether);
        token.mint(alice, 50 ether);
        
        assertEq(token.balanceOf(alice), 150 ether);
        assertEq(token.balanceOf(bob), 200 ether);
        assertEq(token.totalSupply(), 350 ether);
    }
    
    function test_MintOnlyOwner() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, alice));
        token.mint(alice, 100 ether);
    }
    
    function test_Transfer() public {
        token.mint(alice, 100 ether);
        
        vm.prank(alice);
        token.transfer(bob, 30 ether);
        
        assertEq(token.balanceOf(alice), 70 ether);
        assertEq(token.balanceOf(bob), 30 ether);
    }
    
    function test_TransferOwnership() public {
        token.transferOwnership(alice);
        
        assertEq(token.owner(), alice);
        
        // Old owner can't mint
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, owner));
        token.mint(bob, 100 ether);
        
        // New owner can mint
        vm.prank(alice);
        token.mint(bob, 100 ether);
        assertEq(token.balanceOf(bob), 100 ether);
    }
    
    function test_TransferOwnershipZeroAddress() public {
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableInvalidOwner.selector, address(0)));
        token.transferOwnership(address(0));
    }
}
