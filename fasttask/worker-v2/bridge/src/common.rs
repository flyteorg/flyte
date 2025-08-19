use async_channel::Sender;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::Duration;

pub const FAILED: i32 = 7; // 7=retryable 8=permanent
pub const QUEUED: i32 = 3;
pub const RUNNING: i32 = 5;
pub const SUCCEEDED: i32 = 6;

pub struct TaskContext {
    pub kill_tx: Sender<()>,
    pub last_ack_timestamp: u64,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct Task {
    pub cmd: Vec<String>,
    pub additional_distribution: Option<String>,
    pub fast_register_dir: Option<String>,
    pub env_vars: Option<HashMap<String, String>>,
    pub unique_task_id: String, // This maps to the task_id in HeartbeatResponse
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Response {
    pub phase: i32,
    pub reason: Option<String>,
    pub unique_task_id: String, // This maps to the task_id in HeartbeatResponse
    pub executor_corrupt: bool,
}

pub trait ToProstDuration {
    fn to_prost(&self) -> prost_types::Duration;
}

impl ToProstDuration for Duration {
    fn to_prost(&self) -> prost_types::Duration {
        prost_types::Duration {
            seconds: self.as_secs() as i64,
            nanos: self.subsec_nanos() as i32,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_response_serde() {
        let response = Response {
            phase: SUCCEEDED,
            reason: Some("Task completed successfully".to_string()),
            unique_task_id: "test-task-123".to_string(),
            executor_corrupt: false,
        };

        let serialized = bincode::serialize(&response).expect("Failed to serialize");
        let deserialized: Response =
            bincode::deserialize(&serialized).expect("Failed to deserialize");

        assert_eq!(deserialized.phase, response.phase);
        assert_eq!(deserialized.reason, response.reason);
        assert_eq!(deserialized.unique_task_id, response.unique_task_id);
        assert_eq!(deserialized.executor_corrupt, response.executor_corrupt);
    }

    #[test]
    fn test_response_deserialize_failing_bytes() {
        // These are the actual bytes from the log that are failing to deserialize
        // Log line:
        //   Failed to deserialize executor response, bytes b"\x01\x06\0\0\0\0/\0\0\0\0\0\0\0rz4lkmbdlvp5wvnxvp8z-y9n7suu0r72bhdk6e9nxoi6a-0\0"
        let bytes =
            b"\x01\x06\0\0\0\0/\0\0\0\0\0\0\0rz4lkmbdlvp5wvnxvp8z-y9n7suu0r72bhdk6e9nxoi6a-0\0";

        println!("=== DEBUGGING THE ACTUAL ISSUE ===");

        // Let me test if this is actually serialized correctly but with extra data
        // The executor might be sending something wrapped or with additional framing

        // First, let's create the exact same response that should be in those bytes
        let expected_response = Response {
            phase: 6, // SUCCEEDED
            reason: None,
            unique_task_id: "rz4lkmbdlvp5wvnxvp8z-y9n7suu0r72bhdk6e9nxoi6a-0".to_string(),
            executor_corrupt: false, // This should be false based on executor code
        };

        println!("Creating test Response:");
        println!("  phase: {}", expected_response.phase);
        println!("  reason: {:?}", expected_response.reason);
        println!("  unique_task_id: {}", expected_response.unique_task_id);
        println!("  executor_corrupt: {}", expected_response.executor_corrupt);

        let correct_serialization = bincode::serialize(&expected_response).unwrap();
        println!(
            "\nCorrect serialization would be: {:?}",
            correct_serialization
        );
        println!("Length: {}", correct_serialization.len());

        // Now let's see if the issue is that there's a framing/length prefix
        println!("\nReceived bytes: {:?}", bytes);
        println!("Length: {}", bytes.len());

        // Maybe the bytes include some kind of wrapper or have an extra byte?
        // Let's try deserializing from different starting positions
        for start_pos in 0..=2 {
            println!(
                "\n--- Trying to deserialize from position {} ---",
                start_pos
            );
            if start_pos < bytes.len() {
                match bincode::deserialize::<Response>(&bytes[start_pos..]) {
                    Ok(response) => {
                        println!("✓ SUCCESS at position {}: {:?}", start_pos, response);
                        return; // Exit early if we find a working position
                    }
                    Err(e) => {
                        println!("✗ Failed at position {}: {}", start_pos, e);
                    }
                }
            }
        }

        // If that doesn't work, maybe there's corruption or the bytes are truncated
        println!("\n--- Analyzing raw structure ---");
        if bytes.len() >= 4 {
            let first_i32 = i32::from_le_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]);
            println!("First 4 bytes as i32: {}", first_i32);
        }
        if bytes.len() >= 8 {
            let first_i64 = i64::from_le_bytes([
                bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5], bytes[6], bytes[7],
            ]);
            println!("First 8 bytes as i64: {}", first_i64);
        }
    }

    #[test]
    fn test_to_prost_duration_zero() {
        let duration = Duration::from_secs(0);
        let prost_duration = duration.to_prost();

        assert_eq!(prost_duration.seconds, 0);
        assert_eq!(prost_duration.nanos, 0);
    }

    #[test]
    fn test_to_prost_duration_seconds_only() {
        let duration = Duration::from_secs(42);
        let prost_duration = duration.to_prost();

        assert_eq!(prost_duration.seconds, 42);
        assert_eq!(prost_duration.nanos, 0);
    }

    #[test]
    fn test_to_prost_duration_with_nanos() {
        let duration = Duration::new(10, 500_000_000);
        let prost_duration = duration.to_prost();

        assert_eq!(prost_duration.seconds, 10);
        assert_eq!(prost_duration.nanos, 500_000_000);
    }

    #[test]
    fn test_to_prost_duration_millis() {
        let duration = Duration::from_millis(1500);
        let prost_duration = duration.to_prost();

        assert_eq!(prost_duration.seconds, 1);
        assert_eq!(prost_duration.nanos, 500_000_000);
    }

    #[test]
    fn test_to_prost_duration_micros() {
        let duration = Duration::from_micros(1_500_000);
        let prost_duration = duration.to_prost();

        assert_eq!(prost_duration.seconds, 1);
        assert_eq!(prost_duration.nanos, 500_000_000);
    }

    #[test]
    fn test_to_prost_duration_max_nanos() {
        let duration = Duration::new(5, 999_999_999);
        let prost_duration = duration.to_prost();

        assert_eq!(prost_duration.seconds, 5);
        assert_eq!(prost_duration.nanos, 999_999_999);
    }
}
