// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Test} from "forge-std/Test.sol";
import {BatchSettlement} from "../src/BatchSettlement.sol";
import {DisputeGame} from "../src/DisputeGame.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";

/**
 * @title PolymarketCLOBTest
 * @notice Basic integration test for BatchSettlement and DisputeGame contracts
 */
contract PolymarketCLOBTest is Test {
    BatchSettlement public batchSettlement;
    DisputeGame public disputeGame;

    address public mockAllocationManager;
    OperatorSet public testOperatorSet;

    function setUp() public {
        // Setup mock allocation manager
        mockAllocationManager = address(0x1234);

        // Create test operator set
        testOperatorSet = OperatorSet({avs: address(this), id: 1});

        // Deploy contracts
        batchSettlement = new BatchSettlement(
            IAllocationManager(mockAllocationManager),
            testOperatorSet
        );

        disputeGame = new DisputeGame(
            IAllocationManager(mockAllocationManager),
            testOperatorSet,
            address(batchSettlement)
        );
    }

    function testContractDeployment() public view {
        // Verify contracts deployed correctly
        assertEq(
            address(batchSettlement.allocationManager()),
            mockAllocationManager
        );
        assertEq(
            address(disputeGame.allocationManager()),
            mockAllocationManager
        );
        assertEq(disputeGame.batchSettlement(), address(batchSettlement));

        // Check initial state
        assertEq(batchSettlement.totalBatchesSubmitted(), 0);
        assertEq(disputeGame.totalDisputes(), 0);
        assertEq(batchSettlement.getQuorumThreshold(), 6667);
    }

    function testBatchRootTracking() public view {
        bytes32 testRoot = keccak256("test root");

        // Initially, batch should not be submitted
        (bool submitted, uint256 blockNumber) = batchSettlement
            .isBatchSubmitted(testRoot);
        assertFalse(submitted);
        assertEq(blockNumber, 0);
    }

    function testDisputeRootTracking() public view {
        bytes32 testRoot = keccak256("test root");

        // Initially, root should not be disputed
        assertFalse(disputeGame.isRootDisputed(testRoot));

        // Check dispute count
        assertEq(disputeGame.getDisputeCount(), 0);
    }
}
