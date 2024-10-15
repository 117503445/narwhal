// Copyright (c) 2022, Mysten Labs, Inc.
// SPDX-License-Identifier: Apache-2.0
use async_trait::async_trait;
use consensus::ConsensusOutput;
use executor::{ExecutionIndices, ExecutionState, ExecutionStateError};
use thiserror::Error;
use tracing::info;

/// A simple/dumb execution engine.
pub struct SimpleExecutionState;

#[async_trait]
impl ExecutionState for SimpleExecutionState {
    type Transaction = Vec<u8>;
    type Error = SimpleExecutionError;
    type Outcome = Vec<u8>;

    async fn handle_consensus_transaction(
        &self,
        _consensus_output: &ConsensusOutput,
        _execution_indices: ExecutionIndices,
        _transaction: Self::Transaction,
    ) -> Result<Self::Outcome, Self::Error> {
        Ok(Vec::default())
    }

    fn deserialize(bytes: &[u8]) -> Result<Self::Transaction, bincode::Error> {
		info!("ywb enter desrialize");
		info!("Bytes to deserialize: {:?}", bytes);

        // // 检查字节数组是否是有效的 UTF-8 编码
        // match std::str::from_utf8(bytes) {
        //     Ok(valid_str) => {
        //         info!("Valid UTF-8 string: {}", valid_str);
        //         bincode::deserialize(bytes)
        //     },
        //     Err(e) => {
        //         info!("Invalid UTF-8 sequence: {:?}", e);
        //         Err(bincode::Error::new(bincode::ErrorKind::Custom("Invalid UTF-8 sequence".into())))
        //     }
        // }
		Ok(bytes.to_vec())
        // bincode::deserialize(bytes)
    }

    fn ask_consensus_write_lock(&self) -> bool {
        true
    }

    fn release_consensus_write_lock(&self) {}

    async fn load_execution_indices(&self) -> Result<ExecutionIndices, Self::Error> {
        Ok(ExecutionIndices::default())
    }
}

impl Default for SimpleExecutionState {
    fn default() -> Self {
        Self
    }
}

/// A simple/dumb execution error.
#[derive(Debug, Error)]
pub enum SimpleExecutionError {
    #[error("Something went wrong in the authority")]
    ServerError,

    #[error("The client made something bad")]
    ClientError,
}

#[async_trait]
impl ExecutionStateError for SimpleExecutionError {
    fn node_error(&self) -> bool {
        match self {
            Self::ServerError => true,
            Self::ClientError => false,
        }
    }
}
