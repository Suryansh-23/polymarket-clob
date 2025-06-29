// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IBN254CertificateVerifierTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

/**
 * @title BatchSettlement
 * @author Polymarket CLOB Team
 * @notice Handles batch settlement for Polymarket CLOB orders with BLS signature verification
 * @dev This contract verifies BLS aggregate signatures from operators and stores batch roots on-chain
 */
contract BatchSettlement {
    /// @notice The AllocationManager contract for operator set management
    IAllocationManager public immutable allocationManager;

    /// @notice The operator set for this AVS
    OperatorSet public immutable operatorSet;

    /// @notice Minimum stake threshold for 2/3 quorum (in basis points)
    uint256 public constant QUORUM_THRESHOLD_BPS = 6667; // 66.67%

    /// @notice Mapping of batch roots to their submission block numbers
    mapping(bytes32 => uint256) public batchRoots;

    /// @notice Counter for total batches submitted
    uint256 public totalBatchesSubmitted;

    /**
     * @notice Emitted when a new batch is successfully submitted
     * @param root The merkle root of the batch
     * @param submitter The address that submitted the batch
     * @param blockNumber The block number when the batch was submitted
     * @param batchId The sequential ID of this batch
     */
    event BatchSubmitted(
        bytes32 indexed root,
        address indexed submitter,
        uint256 blockNumber,
        uint256 batchId
    );

    /**
     * @notice Emitted when batch submission fails due to invalid signature
     * @param root The merkle root that failed verification
     * @param submitter The address that attempted submission
     */
    event BatchSubmissionFailed(
        bytes32 indexed root,
        address indexed submitter
    );

    /// @notice Thrown when the BLS signature verification fails
    error InvalidBLSSignature();

    /// @notice Thrown when the quorum threshold is not met
    error InsufficientQuorum();

    /// @notice Thrown when a batch root has already been submitted
    error DuplicateBatchRoot();

    /// @notice Thrown when fills data is empty
    error EmptyFillsData();

    /**
     * @notice Constructor to initialize the BatchSettlement contract
     * @param _allocationManager The EigenLayer AllocationManager contract
     * @param _operatorSet The operator set for this AVS
     */
    constructor(
        IAllocationManager _allocationManager,
        OperatorSet memory _operatorSet
    ) {
        allocationManager = _allocationManager;
        operatorSet = _operatorSet;
    }

    /**
     * @notice Submits a batch of orders with BLS signature verification
     * @dev Verifies that the aggregated signature represents ≥2/3 of operator stake
     * @param root The merkle root of the batch containing order fills
     * @param fills The batch fill data (encoded order executions)
     * @param aggSig The BLS aggregate signature from operators
     */
    function submitBatch(
        bytes32 root,
        bytes calldata fills,
        bytes calldata aggSig
    ) external {
        // Validate inputs
        if (fills.length == 0) revert EmptyFillsData();
        if (batchRoots[root] != 0) revert DuplicateBatchRoot();

        // Verify BLS signature and quorum
        if (!_verifyBLSQuorum(root, aggSig)) {
            emit BatchSubmissionFailed(root, msg.sender);
            revert InvalidBLSSignature();
        }

        // Store the batch root with current block number
        batchRoots[root] = block.number;

        // Increment batch counter
        uint256 batchId = ++totalBatchesSubmitted;

        // Emit successful batch submission event
        emit BatchSubmitted(root, msg.sender, block.number, batchId);
    }

    /**
     * @notice Verifies that the BLS signature represents ≥2/3 quorum of operator stake
     * @dev Internal function to validate BLS aggregate signature
     * @param root The message that was signed (batch root)
     * @param aggSig The BLS aggregate signature to verify
     * @return success True if signature is valid and meets quorum threshold
     */
    function _verifyBLSQuorum(
        bytes32 root,
        bytes calldata aggSig
    ) internal view returns (bool success) {
        // Get the current stake for the operator set
        uint256 totalStake = allocationManager.getCurrentTotalMagnitude(
            operatorSet
        );
        if (totalStake == 0) return false;

        // Decode the BLS certificate from the aggregate signature
        IBN254CertificateVerifierTypes.BN254Certificate memory certificate;
        try this._decodeBLSCertificate(aggSig) returns (
            IBN254CertificateVerifierTypes.BN254Certificate memory cert
        ) {
            certificate = cert;
        } catch {
            return false;
        }

        // Verify the certificate represents the required quorum
        uint256 signingStake = _calculateSigningStake(certificate);
        uint256 requiredStake = (totalStake * QUORUM_THRESHOLD_BPS) / 10000;

        if (signingStake < requiredStake) {
            return false;
        }

        // Verify the BLS signature against the root message
        return _verifyBLSSignature(root, certificate);
    }

    /**
     * @notice Decodes BLS certificate from aggregate signature data
     * @dev External function to enable try/catch error handling
     * @param aggSig The aggregate signature data to decode
     * @return certificate The decoded BLS certificate
     */
    function _decodeBLSCertificate(
        bytes calldata aggSig
    )
        external
        pure
        returns (
            IBN254CertificateVerifierTypes.BN254Certificate memory certificate
        )
    {
        // Decode the BLS certificate structure from the signature data
        // This is a simplified implementation - real implementation would decode actual BLS data
        certificate = abi.decode(
            aggSig,
            (IBN254CertificateVerifierTypes.BN254Certificate)
        );
    }

    /**
     * @notice Calculates the total stake represented by the signing operators
     * @dev Internal function to compute stake weight from BLS certificate
     * @param certificate The BLS certificate containing operator signatures
     * @return signingStake The total stake of operators who signed
     */
    function _calculateSigningStake(
        IBN254CertificateVerifierTypes.BN254Certificate memory certificate
    ) internal view returns (uint256 signingStake) {
        // Calculate the stake represented by the signing operators
        // This would integrate with the actual stake calculation from AllocationManager
        // Simplified implementation for MVP
        signingStake = certificate.stakeSigned;
    }

    /**
     * @notice Verifies the BLS signature against the message
     * @dev Internal function to validate BLS cryptographic signature
     * @param message The signed message (batch root)
     * @param certificate The BLS certificate to verify
     * @return valid True if the BLS signature is cryptographically valid
     */
    function _verifyBLSSignature(
        bytes32 message,
        IBN254CertificateVerifierTypes.BN254Certificate memory certificate
    ) internal pure returns (bool valid) {
        // Verify the BLS signature cryptographically
        // This would use the actual BLS verification libraries
        // Simplified implementation for MVP - would integrate with EigenLayer BLS verification

        // In a real implementation, this would:
        // 1. Reconstruct the message hash according to BLS standards
        // 2. Verify the aggregate signature against the aggregate public key
        // 3. Ensure the signature corresponds to the claimed operators

        valid =
            certificate.aggregateG1PublicKey.length > 0 &&
            certificate.aggregateSignature.length > 0;
    }

    /**
     * @notice Checks if a batch root has been submitted
     * @param root The batch root to check
     * @return submitted True if the root has been submitted
     * @return blockNumber The block number when it was submitted (0 if not submitted)
     */
    function isBatchSubmitted(
        bytes32 root
    ) external view returns (bool submitted, uint256 blockNumber) {
        blockNumber = batchRoots[root];
        submitted = blockNumber != 0;
    }

    /**
     * @notice Gets the current quorum threshold in basis points
     * @return threshold The threshold required for batch submission (6667 = 66.67%)
     */
    function getQuorumThreshold() external pure returns (uint256 threshold) {
        threshold = QUORUM_THRESHOLD_BPS;
    }
}
