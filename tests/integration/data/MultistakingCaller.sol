// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.17;

interface IERC20 {
    function transfer(address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

interface IMultiStaking {
    function delegate(
        string memory contractAddress,
        string memory validatorAddress,
        string memory amount
    ) external returns (bool success);
}

contract MultistakingCaller {
    IMultiStaking constant MULTISTAKING =
        IMultiStaking(0x0000000000000000000000000000000000000900);

    /// @notice Delegates through the multistaking precompile, optionally
    /// performing an ERC20 transfer before and/or after the precompile call
    /// within the same transaction.
    function testDelegateWithTransfer(
        address _token,
        string memory _tokenHex,
        address _receiver,
        string memory _validator,
        string memory _delegateAmount,
        uint256 _transferAmount,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            require(
                IERC20(_token).transfer(_receiver, _transferAmount),
                "transfer before delegation failed"
            );
        }

        bool success = MULTISTAKING.delegate(
            _tokenHex,
            _validator,
            _delegateAmount
        );
        require(success, "failed to delegate");

        if (_after) {
            require(
                IERC20(_token).transfer(_receiver, _transferAmount),
                "transfer after delegation failed"
            );
        }
    }
}
