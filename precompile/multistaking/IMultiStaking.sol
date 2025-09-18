// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "./Types.sol";

enum BondStatus { Unbonded, Unbonding, Bonded }

struct Description {
    string moniker;
    string identity;
    string website;
    string securityContact;
    string details;
}

struct Validator {
    string operatorAddress;
    string consensusPubkey;
    bool jailed;
    BondStatus status;
    uint256 tokens;
    uint256 delegatorShares;
    Description description;
    int64 unbondingHeight;
    int64 unbondingTime;
    uint256 commission;
    uint256 minSelfDelegation;
    string bondDenom;
}

struct UnbondingDelegationEntry {
    int64 creationHeight;
    uint256 balance;
}

/// @dev Represents the output of the UnbondingDelegation query.
struct UnbondingDelegationOutput {
    string delegatorAddress;
    string validatorAddress;
    UnbondingDelegationEntry[] entries;
}

interface IMultiStaking {
    function delegate(
        string calldata erc20Token,
        string calldata validatorAddress,
        string calldata amount
    ) external returns (bool success);

    function undelegate(
        string calldata erc20Token,
        string calldata validatorAddress,
        string calldata amount
    ) external returns (int64 completionTime);

    function redelegate(
        string calldata erc20Token,
        string calldata srcValidatorAddress,
        string calldata dstValidatorAddress,
        string calldata amount
    ) external returns (int64 completionTime);

    function cancelUnbondingDelegation(
        string calldata erc20Token,
        string calldata validatorAddress,
        string calldata amount,
        string calldata creationHeight
    ) external returns (bool success);

    function createValidator(
        string calldata pubkey,
        string calldata contractAddress,
        string calldata amount,
        string calldata moniker,
        string calldata identity,
        string calldata website,
        string calldata security,
        string calldata details,
        string calldata commissionRate,
        string calldata commissionMaxRate,
        string calldata commissionMaxChangeRate,
        string calldata minSelfDelegation
    ) external returns (bool success);

    function delegation(
        address delegatorAddress,
        string memory validatorAddress
    ) external view returns (Coin calldata balance);

    function unbondingDelegation(
        address delegatorAddress,
        string memory validatorAddress
    )
        external
        view
        returns (UnbondingDelegationOutput calldata unbondingDelegation);
    
    function validator(
        address validatorAddress
    ) external view returns (Validator calldata validator);
}
