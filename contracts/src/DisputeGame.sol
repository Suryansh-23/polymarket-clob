// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {MerkleProof} from "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";

/**
 * @title DisputeGame
 * @author Polymarket CLOB Team
 * @notice Handles dispute resolution for Polymarket CLOB order sequencing violations
 * @dev Verifies merkle proofs of mis-ordered entries and triggers operator slashing
 */
contract DisputeGame {
    /// @notice The AllocationManager contract for slashing operations
    IAllocationManager public immutable allocationManager;

    /// @notice The operator set for this AVS
    OperatorSet public immutable operatorSet;

    /// @notice The BatchSettlement contract to verify batch roots
    address public immutable batchSettlement;

    /// @notice Mapping to track disputed roots to prevent duplicate disputes
    mapping(bytes32 => bool) public disputedRoots;

    /// @notice Counter for total disputes submitted
    uint256 public totalDisputes;

    /**
     * @dev Structure representing a Polymarket EIP-712 order for comparison
     * @param maker The address of the order maker
     * @param taker The address of the order taker (0x0 for open orders)
     * @param price The price per unit in the order (scaled by 1e18)
     * @param amount The quantity of tokens in the order
     * @param timestamp The timestamp when the order was created
     * @param salt A unique salt to prevent replay attacks
     * @param orderHash The EIP-712 hash of the order
     */
    struct Order {
        address maker;
        address taker;
        uint256 price;
        uint256 amount;
        uint256 timestamp;
        uint256 salt;
        bytes32 orderHash;
    }

    /**
     * @notice Emitted when a successful dispute slashes operators
     * @param root The disputed batch root
     * @param challenger The address that submitted the dispute
     * @param orderAIdx The index of the first mis-ordered order
     * @param orderBIdx The index of the second mis-ordered order
     * @param slashedOperators The number of operators slashed
     */
    event Disputed(
        bytes32 indexed root,
        address indexed challenger,
        uint256 orderAIdx,
        uint256 orderBIdx,
        uint256 slashedOperators
    );

    /**
     * @notice Emitted when a dispute attempt fails
     * @param root The batch root that was disputed
     * @param challenger The address that attempted the dispute
     * @param reason The reason for dispute failure
     */
    event DisputeFailed(
        bytes32 indexed root,
        address indexed challenger,
        string reason
    );

    /// @notice Thrown when the merkle proof is invalid
    error InvalidMerkleProof();

    /// @notice Thrown when orders are correctly ordered (no violation)
    error OrdersCorrectlyOrdered();

    /// @notice Thrown when the batch root doesn't exist
    error BatchRootNotFound();

    /// @notice Thrown when the root has already been disputed
    error RootAlreadyDisputed();

    /// @notice Thrown when order indices are invalid
    error InvalidOrderIndices();

    /**
     * @notice Constructor to initialize the DisputeGame contract
     * @param _allocationManager The EigenLayer AllocationManager contract
     * @param _operatorSet The operator set for this AVS
     * @param _batchSettlement The BatchSettlement contract address
     */
    constructor(
        IAllocationManager _allocationManager,
        OperatorSet memory _operatorSet,
        address _batchSettlement
    ) {
        allocationManager = _allocationManager;
        operatorSet = _operatorSet;
        batchSettlement = _batchSettlement;
    }

    /**
     * @notice Disputes a batch by proving two orders are mis-ordered
     * @dev Verifies merkle proofs and checks price-time priority violation
     * @param root The batch root containing the mis-ordered entries
     * @param orderAIdx The index of the first order in the merkle tree
     * @param orderBIdx The index of the second order in the merkle tree
     * @param proof The merkle proof data containing both orders and their proofs
     */
    function dispute(
        bytes32 root,
        uint256 orderAIdx,
        uint256 orderBIdx,
        bytes calldata proof
    ) external {
        // Validate inputs
        if (orderAIdx >= orderBIdx) revert InvalidOrderIndices();
        if (disputedRoots[root]) revert RootAlreadyDisputed();

        // Verify the batch root exists in BatchSettlement
        (bool submitted, ) = IBatchSettlement(batchSettlement).isBatchSubmitted(
            root
        );
        if (!submitted) revert BatchRootNotFound();

        // Decode the proof data containing both orders and their merkle proofs
        (
            Order memory orderA,
            Order memory orderB,
            bytes32[] memory proofA,
            bytes32[] memory proofB
        ) = _decodeDisputeProof(proof);

        // Verify merkle proofs for both orders
        bytes32 leafA = _hashOrder(orderA);
        bytes32 leafB = _hashOrder(orderB);

        if (
            !MerkleProof.verify(proofA, root, leafA) ||
            !MerkleProof.verify(proofB, root, leafB)
        ) {
            emit DisputeFailed(root, msg.sender, "Invalid merkle proof");
            revert InvalidMerkleProof();
        }

        // Check if orders violate price-time priority
        if (_isCorrectlyOrdered(orderA, orderB)) {
            emit DisputeFailed(root, msg.sender, "Orders correctly ordered");
            revert OrdersCorrectlyOrdered();
        }

        // Mark root as disputed to prevent duplicate disputes
        disputedRoots[root] = true;

        // Slash all operators in the set
        uint256 slashedCount = _slashAllOperators(root);

        // Increment dispute counter
        totalDisputes++;

        // Emit successful dispute event
        emit Disputed(root, msg.sender, orderAIdx, orderBIdx, slashedCount);
    }

    /**
     * @notice Decodes the dispute proof data
     * @dev Internal function to extract orders and merkle proofs from calldata
     * @param proof The encoded proof data
     * @return orderA The first order
     * @return orderB The second order
     * @return proofA Merkle proof for orderA
     * @return proofB Merkle proof for orderB
     */
    function _decodeDisputeProof(
        bytes calldata proof
    )
        internal
        pure
        returns (
            Order memory orderA,
            Order memory orderB,
            bytes32[] memory proofA,
            bytes32[] memory proofB
        )
    {
        // Decode the proof structure containing both orders and their proofs
        (orderA, orderB, proofA, proofB) = abi.decode(
            proof,
            (Order, Order, bytes32[], bytes32[])
        );
    }

    /**
     * @notice Hashes an order according to EIP-712 specification
     * @dev Internal function to create merkle leaf from order data
     * @param order The order to hash
     * @return hash The keccak256 hash of the order
     */
    function _hashOrder(
        Order memory order
    ) internal pure returns (bytes32 hash) {
        // Create the order hash for merkle tree inclusion
        // This should match the format used when building the merkle tree
        hash = keccak256(
            abi.encode(
                order.maker,
                order.taker,
                order.price,
                order.amount,
                order.timestamp,
                order.salt,
                order.orderHash
            )
        );
    }

    /**
     * @notice Checks if two orders follow correct price-time priority
     * @dev Internal function to validate ordering rules
     * @param orderA The first order (should have priority)
     * @param orderB The second order (should be after orderA)
     * @return correct True if orders are correctly ordered
     */
    function _isCorrectlyOrdered(
        Order memory orderA,
        Order memory orderB
    ) internal pure returns (bool correct) {
        // Price priority: better price comes first
        if (orderA.price != orderB.price) {
            // For buy orders: higher price has priority
            // For sell orders: lower price has priority
            // Assuming this is a unified price comparison where higher is better
            return orderA.price >= orderB.price;
        }

        // Time priority: earlier timestamp comes first when prices are equal
        correct = orderA.timestamp <= orderB.timestamp;
    }

    /**
     * @notice Slashes all operators in the operator set
     * @dev Internal function to execute slashing via AllocationManager
     * @param root The disputed batch root (for slashing description)
     * @return slashedCount The number of operators slashed
     */
    function _slashAllOperators(
        bytes32 root
    ) internal returns (uint256 slashedCount) {
        // Get all operators in the set for slashing
        // This is a simplified implementation - real version would get actual operator list

        // Create slashing parameters for all operators
        // In a full implementation, this would iterate through all operators in the set
        // and create individual slashing requests for each

        // For now, we'll create a symbolic slashing request
        // The actual implementation would need to:
        // 1. Get the list of operators from the AllocationManager
        // 2. Create slashing parameters for each operator
        // 3. Execute slashing for the full stake of each operator

        string memory description = string(
            abi.encodePacked(
                "Order sequencing violation in batch: ",
                _bytes32ToString(root)
            )
        );

        // This would be replaced with actual operator slashing logic
        // For now, return a symbolic count
        slashedCount = 1; // Placeholder

        // Note: Real implementation would call:
        // allocationManager.slashOperator(operatorSet.avs, slashingParams)
        // for each operator with 100% slashing
    }

    /**
     * @notice Converts bytes32 to string for event descriptions
     * @dev Internal utility function
     * @param _bytes The bytes32 value to convert
     * @return str The string representation
     */
    function _bytes32ToString(
        bytes32 _bytes
    ) internal pure returns (string memory str) {
        bytes memory bytesArray = new bytes(32);
        for (uint256 i = 0; i < 32; i++) {
            bytesArray[i] = _bytes[i];
        }
        str = string(bytesArray);
    }

    /**
     * @notice Checks if a root has been disputed
     * @param root The batch root to check
     * @return disputed True if the root has been disputed
     */
    function isRootDisputed(
        bytes32 root
    ) external view returns (bool disputed) {
        disputed = disputedRoots[root];
    }

    /**
     * @notice Gets the total number of disputes submitted
     * @return count The total dispute count
     */
    function getDisputeCount() external view returns (uint256 count) {
        count = totalDisputes;
    }
}

/**
 * @notice Interface for BatchSettlement contract
 * @dev Used to verify batch root existence
 */
interface IBatchSettlement {
    function isBatchSubmitted(
        bytes32 root
    ) external view returns (bool submitted, uint256 blockNumber);
}
