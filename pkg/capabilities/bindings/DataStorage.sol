// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

contract DataStorage {
    // Mapping to store data keyed by an address and a string key
    mapping(address => mapping(string => string)) private data;

    // Event emitted when data is stored
    event DataStored(address indexed sender, string key, string value);

    // New event emitted by a different method
    event AccessLogged(address indexed caller, string message);

    // Custom error for when a key is not found
    error DataNotFound(address requester, string key, string reason);

    // Struct definition
    struct UserData {
        string key;
        string value;
    }

    // Write method: Stores a key-value pair
    function storeData(string calldata key, string calldata value) external {
        data[msg.sender][key] = value;
        emit DataStored(msg.sender, key, value);
    }

    // Read method: Retrieves the value for a given key
    // No longer emits any event
    function readData(address user, string calldata key) external view returns (string memory) {
        string memory value = data[user][key];

        if (bytes(value).length == 0) {
            revert DataNotFound(user, key, "No data associated with this key.");
        }

        return value;
    }

    // New method: Emits a different event
    function logAccess(string calldata message) external {
        emit AccessLogged(msg.sender, message);
    }

    // New method: Accepts a struct and stores its data
    function storeUserData(UserData calldata userData) external {
        data[msg.sender][userData.key] = userData.value;
        emit DataStored(msg.sender, userData.key, userData.value);
    }

    function onReport(bytes calldata metadata, bytes calldata payload) external {
        UserData memory user = abi.decode(payload, (UserData));
        // TODO implement logic to handle the report
        data[msg.sender][user.key] = user.value;
        emit DataStored(msg.sender, user.key, user.value);
    }
}