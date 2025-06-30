// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Script, console} from "forge-std/Script.sol";
import {stdJson} from "forge-std/StdJson.sol";

import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {IKeyRegistrar} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {IBN254CertificateVerifier} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {ITaskMailbox} from "@hourglass-monorepo/src/interfaces/core/ITaskMailbox.sol";

import {TaskAVSRegistrar} from "@project/l1-contracts/TaskAVSRegistrar.sol";
import {AVSTaskHook} from "@project/l2-contracts/AVSTaskHook.sol";
import {BatchSettlement} from "@project/BatchSettlement.sol";
import {DisputeGame} from "@project/DisputeGame.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

/**
 * @title DeployPolymarketAVS
 * @author Polymarket CLOB Team
 * @notice Deployment script for Polymarket CLOB AVS contracts (BatchSettlement and DisputeGame)
 * @dev This script deploys the core AVS contracts with proper dependency management and validation
 */
contract DeployPolymarketAVS is Script {
    using stdJson for string;

    /// @notice Configuration for operator set ID (can be overridden via environment)
    uint32 public constant DEFAULT_OPERATOR_SET_ID = 0;

    /// @notice Deployment configuration and context
    struct Context {
        address avs;
        uint256 avsPrivateKey;
        uint256 deployerPrivateKey;
        IAllocationManager allocationManager;
        IKeyRegistrar keyRegistrar;
        IBN254CertificateVerifier certificateVerifier;
        ITaskMailbox taskMailbox;
        TaskAVSRegistrar taskAVSRegistrar;
        AVSTaskHook taskHook;
    }

    /// @notice Output structure for deployed contracts
    struct Output {
        string name;
        address contractAddress;
    }

    /// @notice Deployment results for external consumption
    struct DeploymentResult {
        address batchSettlement;
        address disputeGame;
        uint256 blockNumber;
        uint256 timestamp;
    }

    /// @notice Emitted when AVS contracts are successfully deployed
    event AVSDeployed(
        address indexed batchSettlement,
        address indexed disputeGame,
        address indexed avs,
        uint32 operatorSetId,
        uint256 blockNumber
    );

    /// @notice Thrown when context validation fails
    error InvalidContext(string reason);

    /// @notice Thrown when contract deployment fails
    error DeploymentFailed(string contractName, string reason);

    /**
     * @notice Main deployment function for Polymarket CLOB AVS contracts
     * @param environment The deployment environment (e.g., "devnet", "testnet", "mainnet")
     * @param _context JSON string containing deployment context and configuration
     */
    function run(
        string memory environment,
        string memory _context
    ) public returns (DeploymentResult memory) {
        console.log("Starting Polymarket CLOB AVS deployment...");
        console.log("Environment:", environment);
        console.log("Block number:", block.number);
        console.log("Chain ID:", block.chainid);

        // Read and validate the deployment context
        Context memory context = _readContext(environment, _context);
        _validateContext(context);

        // Get operator set ID from environment or use default
        uint32 operatorSetId = uint32(
            vm.envOr("OPERATOR_SET_ID", uint256(DEFAULT_OPERATOR_SET_ID))
        );
        console.log("Using operator set ID:", operatorSetId);

        DeploymentResult memory result;

        vm.startBroadcast(context.deployerPrivateKey);
        console.log("Deployer address:", vm.addr(context.deployerPrivateKey));

        // Create operator set for the AVS
        OperatorSet memory operatorSet = OperatorSet({
            avs: context.avs,
            id: operatorSetId
        });

        console.log("Deploying with operator set:");
        console.log("  AVS address:", operatorSet.avs);
        console.log("  Operator set ID:", operatorSet.id);

        // Deploy BatchSettlement contract first (no dependencies)
        console.log("Deploying BatchSettlement...");
        BatchSettlement batchSettlement = new BatchSettlement(
            context.allocationManager,
            operatorSet
        );

        if (address(batchSettlement) == address(0)) {
            revert DeploymentFailed(
                "BatchSettlement",
                "Contract address is zero"
            );
        }

        console.log("BatchSettlement deployed to:", address(batchSettlement));
        result.batchSettlement = address(batchSettlement);

        // Verify BatchSettlement deployment
        uint256 threshold = batchSettlement.getQuorumThreshold();
        console.log(
            "BatchSettlement verification: quorum threshold =",
            threshold
        );

        // Deploy DisputeGame contract with BatchSettlement dependency
        console.log("Deploying DisputeGame...");
        DisputeGame disputeGame = new DisputeGame(
            context.allocationManager,
            operatorSet,
            address(batchSettlement)
        );

        if (address(disputeGame) == address(0)) {
            revert DeploymentFailed("DisputeGame", "Contract address is zero");
        }

        console.log("DisputeGame deployed to:", address(disputeGame));
        result.disputeGame = address(disputeGame);

        // Verify DisputeGame deployment and BatchSettlement linkage
        address linkedBatchSettlement = disputeGame.batchSettlement();
        if (linkedBatchSettlement != address(batchSettlement)) {
            revert DeploymentFailed(
                "DisputeGame",
                "BatchSettlement linkage verification failed"
            );
        }
        console.log(
            "DisputeGame verification: linked to BatchSettlement =",
            linkedBatchSettlement
        );

        result.blockNumber = block.number;
        result.timestamp = block.timestamp;

        console.log("Core contracts deployment completed successfully!");

        vm.stopBroadcast();

        vm.startBroadcast(context.avsPrivateKey);
        console.log("AVS address:", context.avs);

        // Additional AVS setup can be added here if needed
        // For example: setting initial parameters, registering operators, etc.
        console.log("Performing additional AVS setup...");

        vm.stopBroadcast();

        // Emit deployment event for external monitoring
        emit AVSDeployed(
            result.batchSettlement,
            result.disputeGame,
            context.avs,
            operatorSetId,
            block.number
        );

        // Write deployment addresses to output file
        Output[] memory outputs = new Output[](2);
        outputs[0] = Output({
            name: "BatchSettlement",
            contractAddress: result.batchSettlement
        });
        outputs[1] = Output({
            name: "DisputeGame",
            contractAddress: result.disputeGame
        });
        _writeOutputToJson(environment, outputs);

        console.log("Deployment summary:");
        console.log("  BatchSettlement:", result.batchSettlement);
        console.log("  DisputeGame:", result.disputeGame);
        console.log("  Block number:", result.blockNumber);
        console.log("  Timestamp:", result.timestamp);
        console.log("Polymarket CLOB AVS deployment completed successfully!");

        return result;
    }

    /**
     * @notice Validates the deployment context to ensure all required fields are present
     * @param context The deployment context to validate
     */
    function _validateContext(Context memory context) internal pure {
        if (context.avs == address(0)) {
            revert InvalidContext("AVS address cannot be zero");
        }
        if (context.avsPrivateKey == 0) {
            revert InvalidContext("AVS private key cannot be zero");
        }
        if (context.deployerPrivateKey == 0) {
            revert InvalidContext("Deployer private key cannot be zero");
        }
        if (address(context.allocationManager) == address(0)) {
            revert InvalidContext("AllocationManager address cannot be zero");
        }
        // Note: KeyRegistrar can be zero for devnet, but should be set for production
        if (address(context.certificateVerifier) == address(0)) {
            revert InvalidContext("CertificateVerifier address cannot be zero");
        }
        // Note: TaskMailbox, TaskAVSRegistrar, and TaskHook can be zero if not deployed yet
    }

    /**
     * @notice Reads and parses the deployment context from JSON configuration
     * @param _context JSON string containing deployment context
     * @return context Parsed deployment context
     */
    function _readContext(
        string memory /* environment */,
        string memory _context
    ) internal pure returns (Context memory) {
        // Parse the context
        Context memory context;
        context.avs = stdJson.readAddress(_context, ".context.avs.address");
        context.avsPrivateKey = uint256(
            stdJson.readBytes32(_context, ".context.avs.avs_private_key")
        );
        context.deployerPrivateKey = uint256(
            stdJson.readBytes32(_context, ".context.deployer_private_key")
        );
        context.allocationManager = IAllocationManager(
            stdJson.readAddress(
                _context,
                ".context.eigenlayer.l1.allocation_manager"
            )
        );

        // KeyRegistrar is not in the devnet.yaml, so we'll use a default address
        // In production, this should be properly configured
        context.keyRegistrar = IKeyRegistrar(
            address(0x1C84Bb62fE7791e173014A879C706445fa893BbE)
        );

        context.certificateVerifier = IBN254CertificateVerifier(
            stdJson.readAddress(
                _context,
                ".context.eigenlayer.l2.bn254_certificate_verifier"
            )
        );

        // Read deployed contract addresses directly from the context
        context.taskMailbox = ITaskMailbox(
            _readDeployedContractAddress(_context, "taskMailbox")
        );
        context.taskAVSRegistrar = TaskAVSRegistrar(
            _readDeployedContractAddress(_context, "taskAVSRegistrar")
        );
        context.taskHook = AVSTaskHook(
            _readDeployedContractAddress(_context, "avsTaskHook")
        );

        return context;
    }

    /**
     * @notice Reads deployed contract address from the context
     * @param contractName Name of the contract to find
     * @return contractAddress The address of the deployed contract
     */
    function _readDeployedContractAddress(
        string memory /* _context */,
        string memory contractName
    ) internal pure returns (address contractAddress) {
        // Use default addresses from the YAML configuration
        // In a production environment, these should be read dynamically

        // For taskMailbox (index 0)
        if (
            keccak256(abi.encodePacked(contractName)) ==
            keccak256(abi.encodePacked("taskMailbox"))
        ) {
            return address(0x0c8A72d89AA40B71Ee5F430E89E1681f944DAAc0); // Default from YAML
        }

        // For taskAVSRegistrar (index 1)
        if (
            keccak256(abi.encodePacked(contractName)) ==
            keccak256(abi.encodePacked("taskAVSRegistrar"))
        ) {
            return address(0xC2100780F828810bAaeCD3039111006E12B26CF8); // Default from YAML
        }

        // For avsTaskHook (index 2)
        if (
            keccak256(abi.encodePacked(contractName)) ==
            keccak256(abi.encodePacked("avsTaskHook"))
        ) {
            return address(0x6b165dE91F825584117c23f79F634849aCff4f68); // Default from YAML
        }

        return address(0);
    }

    /**
     * @notice Writes deployment output to JSON file for external consumption
     * @param environment The deployment environment
     * @param outputs Array of deployed contract outputs
     */
    function _writeOutputToJson(
        string memory environment,
        Output[] memory outputs
    ) internal {
        uint256 length = outputs.length;

        if (length > 0) {
            // Add the addresses object
            string memory addresses = "addresses";

            for (uint256 i = 0; i < outputs.length - 1; i++) {
                vm.serializeAddress(
                    addresses,
                    outputs[i].name,
                    outputs[i].contractAddress
                );
            }
            addresses = vm.serializeAddress(
                addresses,
                outputs[length - 1].name,
                outputs[length - 1].contractAddress
            );

            // Add the chainInfo object
            string memory chainInfo = "chainInfo";
            chainInfo = vm.serializeUint(chainInfo, "chainId", block.chainid);

            // Finalize the JSON
            string memory finalJson = "final";
            vm.serializeString(finalJson, "addresses", addresses);
            finalJson = vm.serializeString(finalJson, "chainInfo", chainInfo);

            // Write to output file
            string memory outputFile = string.concat(
                "script/",
                environment,
                "/output/deploy_polymarket_avs_output.json"
            );
            vm.writeJson(finalJson, outputFile);
        }
    }
}
