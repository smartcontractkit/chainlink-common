// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IReserveManager {
    function updateReserves(uint256 totalMinted, uint256 totalReserve) external;
}
