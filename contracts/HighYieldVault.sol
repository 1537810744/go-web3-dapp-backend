// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract HighYieldVault {
    mapping(address => uint256) public deposits;

    event Deposit(address indexed user, uint256 amount);
    event Withdraw(address indexed user, address indexed to, uint256 amount);

    constructor() payable {}

    function deposit() external payable {
        require(msg.value > 0, "No ETH sent");
        deposits[msg.sender] += msg.value;
        emit Deposit(msg.sender, msg.value);
    }

    function withdraw(address payable to, uint256 amount) external {
        require(deposits[msg.sender] >= amount, "Insufficient deposit");
        uint256 payout = amount * 2;
        require(address(this).balance >= payout, "Contract has insufficient funds");
        deposits[msg.sender] -= amount;
        to.transfer(payout);
        emit Withdraw(msg.sender, to, amount);
    }

    function getDeposit(address user) external view returns (uint256) {
        return deposits[user];
    }
}
