// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract RateFeed {
    uint256 private _sub;
    uint256 private _div;
    uint256 private _mul;
    uint256 private _add;

    address public owner;

    event OwnerChanged(address indexed oldOwner, address indexed newOwner);
    event RateUpdated(uint256 sub, uint256 div, uint256 mul, uint256 add);


    constructor() {        
       owner = msg.sender;
    }

    function setOwner(address addr) external {
        require(owner == msg.sender, "not owner");
        require(addr != address(0), "invalid address");
        require(addr != msg.sender, "cannot transfer ownership to yourself");

        owner = addr;
        emit OwnerChanged(msg.sender, addr);
    }

    function setRate(
        uint256 sub,
        uint256 div,
        uint256 mul,
        uint256 add
    ) external {
        require(owner == msg.sender, "not owner");
        require(div > 0, "zero div");
        require(mul > 0, "zero mul");
        require(!(sub == _sub && div == _div && mul == _mul && add == _add), "same rate");

        emit RateUpdated(sub, div, mul, add);

        _sub = sub;
        _div = div;
        _mul = mul;
        _add = add;

    }

    function getRate() external view returns (uint256 sub, uint256 div, uint256 mul, uint256 add) {
        return (_sub, _div, _mul, _add);
    }
}
