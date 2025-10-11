// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract MultiOwnerValues {
    uint256[4] private feed;

    mapping(address => bool) private owners;

    event OwnerAdded(address indexed newOwner);
    event ValuesUpdated(uint256 sub, uint256 div, uint256 mul, uint256 add);

    modifier onlyOwner() {
        require(owners[msg.sender], "Not an owner");
        _;
    }

    constructor() {        
       owners[msg.sender] = true;
    }

    function _addOwner(address _owner) internal {
        require(_owner != address(0), "Invalid address");
        require(!owners[_owner], "Already an owner");
        owners[_owner] = true;
        emit OwnerAdded(_owner);
    }

    function addOwner(address _newOwner) external onlyOwner {
        _addOwner(_newOwner);
    }

    function setFeed(
        uint256 sub,
        uint256 div,
        uint256 mul,
        uint256 add
    ) external onlyOwner {
        feed[0] = sub;
        feed[1] = div;
        feed[2] = mul;
        feed[3] = add;
        emit ValuesUpdated(sub, div, mul, add);
    }

    function getValues() external view returns (uint256[4] memory) {
        return feed;
    }

    function isOwner(address _addr) external view returns (bool) {
        return owners[_addr];
    }
}
