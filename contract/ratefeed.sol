// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract MultiOwnerValues {
    uint256 public sub;
    uint256 public div;
    uint256 public mul;
    uint256 public add;

    address public owner;

    event OwnerChanged(address indexed oldOwner, address indexed newOwner);
    event RateUpdated(uint256 sub, uint256 div, uint256 mul, uint256 add);

    modifier onlyOwner() {
        require(owner == msg.sender, "Not an owner");
        _;
    }

    constructor() {        
       owner = msg.sender;
    }

    function setOwner(address _addr) external onlyOwner {
        require(_addr != address(0), "Invalid address");
        require(_addr != msg.sender, "cannot transfer ownership to yourself");

        owner = _addr;
        emit OwnerChanged(msg.sender, _addr);
    }

    function setRate(
        uint256 _sub,
        uint256 _div,
        uint256 _mul,
        uint256 _add
    ) external onlyOwner {
        require(_div > 0, "div cannot be zero");
        require(_mul > 0, "mul cannot be zero");
        require(!(sub == _sub && div == _div && mul == _mul && add == _add), "rate unchanged");

        sub = _sub;
        div = _div;
        mul = _mul;
        add = _add;
        emit RateUpdated(sub, div, mul, add);
    }

    function getRate() external view returns (uint256 _sub, uint256 _div, uint256 _mul, uint256 _add) {
        return (sub, div, mul, add);
    }
}
