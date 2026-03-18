// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

interface IFeeGrant {
    /// @dev Grant a fee allowance from granter to grantee.
    /// Determines allowance type based on parameters:
    /// - If period > 0 and periodLimit is non-empty: creates PeriodicAllowance
    /// - Otherwise: creates BasicAllowance
    /// - If allowedMessages is non-empty: wraps in AllowedMsgAllowance
    /// @param grantee The address of the grantee
    /// @param spendLimit The maximum amount of coins that can be spent as a comma-separated string (e.g. "100stake", empty for unlimited)
    /// @param expiration The expiration time in RFC3339 format (e.g. "2026-01-01T00:00:00Z", empty for no expiration)
    /// @param period The period duration in seconds (0 for no periodic allowance)
    /// @param periodLimit The maximum amount of coins per period as a comma-separated string (e.g. "10stake", empty if no periodic allowance)
    /// @param allowedMessages The list of allowed message types (empty for no restriction)
    function grant(
        address grantee,
        string calldata spendLimit,
        string calldata expiration,
        int64 period,
        string calldata periodLimit,
        string[] calldata allowedMessages
    ) external returns (bool success);

    /// @dev Revoke a fee allowance from granter to grantee
    /// @param grantee The address of the grantee
    function revoke(
        address grantee
    ) external returns (bool success);
}

