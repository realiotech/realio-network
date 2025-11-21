// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

/// @dev The IBank contract's address.
address constant IBANK_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000804;

/// @dev The IBank contract's instance.
IBank constant IBANK_CONTRACT = IBank(IBANK_PRECOMPILE_ADDRESS);

/// @dev Balance specifies the denom and the amount of tokens.
struct Balance {
    /// denom of tokens
    string denom;
    /// amount of tokens
    uint256 amount;
}

struct Input {
    address addr;
    string denom;
    uint256 amount;
}

struct Output {
    address addr;
    string amount;
}

/**
 * @author Evmos Team
 * @title Bank Interface
 * @dev Interface for querying balances and supply from the Bank module.
 */
interface IBank {
    /// @dev balances defines a method for retrieving all the native token balances
    /// for a given account.
    /// @param account the address of the account to query balances for.
    /// @return balances the array of native token balances.
    function balances(
        address account
    ) external view returns (Balance[] memory balances);

    /// @dev totalSupply defines a method for retrieving the total supply of all
    /// native tokens.
    /// @return totalSupply the supply as an array of native token balances
    function totalSupply() external view returns (Balance[] memory totalSupply);

    /// @dev supplyOf defines a method for retrieving the total supply of a particular native coin.
    /// @return totalSupply the supply as a uint256
    function supplyOf(
        string memory denom
    ) external view returns (uint256 totalSupply);

    function send(
        address to,
        string memory coins
    ) external view returns (bool success);

    function multiSend(
        string memory coins,
        Output[] calldata output
    ) external view returns (bool success);
}
