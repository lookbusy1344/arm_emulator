/**
 * Centralized test constants to eliminate magic numbers
 */

// Timeout constants (in milliseconds)
export const TIMEOUTS = {
  // Short delays for UI updates
  UI_UPDATE: 50,
  UI_STABILIZE: 100,

  // VM state changes
  VM_STATE_CHANGE: 200,
  VM_RESET: 300,

  // Execution timeouts
  EXECUTION_START: 300,
  EXECUTION_SHORT: 1000,
  EXECUTION_NORMAL: 5000,
  EXECUTION_MAX: 10000,

  // Step operation
  STEP_COMPLETE: 100,
  STEP_OVER_COMPLETE: 500,
} as const;

// Memory addresses
export const ADDRESSES = {
  CODE_SEGMENT_START: 0x00008000,
  STACK_START: 0x00050000,
  NULL: 0x00000000,
} as const;

// Expected arithmetic results (from arithmetic test program)
export const ARITHMETIC_RESULTS = {
  ADD_10_20: 0x0000001E,  // 30 in hex (10 + 20)
  SUB_20_10: 0x0000000A,  // 10 in hex (20 - 10)
  MUL_10_20: 0x000000C8,  // 200 in hex (10 * 20)
} as const;

// Wails backend configuration
export const BACKEND = {
  PORT: 34115,
  BASE_URL: 'http://localhost:34115',
  HEALTH_CHECK_INTERVAL: 2000,  // ms between health checks
  HEALTH_CHECK_MAX_WAIT: 120000,  // 2 minutes max wait
} as const;

// Test execution limits
export const LIMITS = {
  MAX_STEPS: 100,
  MAX_LOOP_ITERATIONS: 1000,
} as const;

// Register names for programmatic access
export const REGISTERS = {
  GENERAL: ['R0', 'R1', 'R2', 'R3', 'R4', 'R5', 'R6', 'R7',
            'R8', 'R9', 'R10', 'R11', 'R12'] as const,
  SPECIAL: {
    STACK_POINTER: 'R13',
    LINK_REGISTER: 'R14',
    PROGRAM_COUNTER: 'PC',
    SP: 'SP',
    LR: 'LR',
  } as const,
} as const;

// Execution states
export const EXECUTION_STATES = {
  IDLE: 'idle',
  RUNNING: 'running',
  PAUSED: 'paused',
  HALTED: 'halted',
  EXITED: 'exited',
  ERROR: 'error',
} as const;

// Visual regression settings
export const VISUAL = {
  // Strict tolerance for catching regressions
  MAX_DIFF_PIXEL_RATIO: 0.02,  // 2% pixel difference
  THRESHOLD: 0.1,  // 10% color difference per pixel

  // Current lenient settings (for comparison)
  CURRENT_MAX_DIFF: 0.06,  // 6%
  CURRENT_THRESHOLD: 0.2,  // 20%
} as const;
