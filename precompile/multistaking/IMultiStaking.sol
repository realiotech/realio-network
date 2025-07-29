// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

enum BondStatus { Unbonded, Unbonding, Bonded }

struct RedelegationEntry {
    int64 creationHeight;
    int64 completionTime;
    uint256 initialBalance;
    uint256 sharesDst;
}

struct RedelegationOutput {
    string delegatorAddress;
    string validatorSrcAddress;
    string validatorDstAddress;
    RedelegationEntry[] entries;
}

struct Redelegation {
    string delegatorAddress;
    string validatorSrcAddress;
    string validatorDstAddress;
    RedelegationEntry[] entries;
}

struct RedelegationEntryResponse {
    RedelegationEntry redelegationEntry;
    uint256 balance;
}

struct RedelegationResponse {
    Redelegation redelegation;
    RedelegationEntryResponse[] entries;
}

struct PageRequest {
    bytes key;
    uint64 offset;
    uint64 limit;
    bool countTotal;
    bool reverse;
}

struct PageResponse {
    bytes nextKey;
    uint64 total;
}

struct Validator {
    string operatorAddress;
    string consensusPubkey;
    bool jailed;
    BondStatus status;
    uint256 tokens;
    uint256 delegatorShares;
    string description;
    int64 unbondingHeight;
    int64 unbondingTime;
    uint256 commission;
    uint256 minSelfDelegation;
}

interface IMultiStaking {
    function delegate(
        string calldata erc20Token,
        string calldata validatorAddress,
        string calldata amount
    ) external returns (bool success);

    function undelegate(
        address erc20Token,
        string calldata validatorAddress,
        uint256 amount
    ) external returns (int64 completionTime);

    function redelegate(
        address erc20Token,
        string calldata srcValidatorAddress,
        string calldata dstValidatorAddress,
        uint256 amount
    ) external returns (int64 completionTime);

    function cancelUnbondingDelegation(
        address erc20Token,
        string calldata validatorAddress,
        uint256 amount,
        int64 creationHeight
    ) external returns (bool success);

    function createValidator(
        string calldata validatorAddress,
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

    function editValidator(
        string calldata validatorAddress,
        string calldata moniker,
        string calldata identity,
        string calldata website,
        string calldata security,
        string calldata details,
        string calldata commissionRate,
        string calldata minSelfDelegation
    ) external returns (bool success);

    function delegation(
        string calldata delegatorAddress,
        string calldata validatorAddress
    ) external view returns (uint256 shares, uint256 balance);

    function unbondingDelegation(
        string calldata delegatorAddress,
        string calldata validatorAddress
    ) external view returns (uint256 balance, int64 completionTime);

    function validator(
        address validatorAddress
    ) external view returns (Validator memory);

    function validators(
        string calldata status,
        PageRequest calldata pageRequest
    ) external view returns (Validator[] memory, PageResponse memory);

    function redelegation(
        address delegatorAddress,
        string calldata srcValidatorAddress,
        string calldata dstValidatorAddress
    ) external view returns (RedelegationOutput memory);

    function redelegations(
        address delegatorAddress,
        string calldata srcValidatorAddress,
        string calldata dstValidatorAddress,
        PageRequest calldata pageRequest
    ) external view returns (RedelegationResponse[] memory, PageResponse memory);
}
