#!/bin/zsh
set -euo pipefail

if (( $# != 2 )); then
  print -u2 -- "usage: $0 LABEL TARGET_DIRECTORY_NAME"
  exit 64
fi

LABEL="$1"
TARGET_NAME="$2"
TASK_ID="TASK-260713-3mlifc"
TASK_ROOT="/Users/alexis/src/relux-works/relux-agents-infra/.temp/${TASK_ID}"
ALIAS_ROOT="/tmp/${TASK_ID}-primary-session"
BIN="${ALIAS_ROOT}/bin/agents-infra"
RUNTIME_SOURCE="${ALIAS_ROOT}/runtime-source"
TARGET="${ALIAS_ROOT}/copies/${TARGET_NAME}"
HOME_DIR="${ALIAS_ROOT}/home/${LABEL}"
FAKE_BIN="${ALIAS_ROOT}/fake-bin"
RESULT_DIR="${TASK_ROOT}/results/${LABEL}"
TRANSCRIPT="${RESULT_DIR}/${TASK_ID}_${LABEL}-transcript.log"
ASSERTIONS="${RESULT_DIR}/${TASK_ID}_${LABEL}-assertions.log"
ROOT_CONFIG="${TARGET}/.agents/.configs/project-config.toml"
CHILD="${TARGET}/.validation/${TASK_ID}/child"
CHILD_CONFIG="${CHILD}/.agents/.configs/project-config.toml"
BOARD_CONFIG="${TARGET}/task-board.config.json"
ORIGINAL_ROOT_CONFIG="${RESULT_DIR}/original-project-config.toml"
ORIGINAL_BOARD_CONFIG="${RESULT_DIR}/original-task-board.config.json"
LAST_LOG=""

mkdir -p "$RESULT_DIR" "$HOME_DIR" "${CHILD}/.agents/.configs"
: > "$TRANSCRIPT"
: > "$ASSERTIONS"

cp "$ROOT_CONFIG" "$ORIGINAL_ROOT_CONFIG"
cp "$BOARD_CONFIG" "$ORIGINAL_BOARD_CONFIG"

cleanup() {
  mkdir -p "${TARGET}/.agents/.configs"
  cp "$ORIGINAL_ROOT_CONFIG" "$ROOT_CONFIG"
  rm -rf "${TARGET}/.validation/${TASK_ID}"
}
trap cleanup EXIT

pass() {
  print -r -- "PASS\t$1" >> "$ASSERTIONS"
}

fail() {
  print -u2 -r -- "FAIL [$LABEL] $1"
  print -r -- "FAIL\t$1" >> "$ASSERTIONS"
  exit 1
}

record_command_line() {
  print -nr -- '$'
  local argument
  for argument in "$@"; do
    printf ' %q' "$argument"
  done
  print
}

run_ok() {
  local name="$1"
  shift
  local log="${RESULT_DIR}/${name}.log"
  record_command_line "$@" >> "$TRANSCRIPT"
  local exit_code
  if "$@" > "$log" 2>&1; then
    exit_code=0
  else
    exit_code=$?
  fi
  cat "$log" >> "$TRANSCRIPT"
  print -r -- "[exit=${exit_code}]" >> "$TRANSCRIPT"
  print >> "$TRANSCRIPT"
  (( exit_code == 0 )) || fail "${name}: expected exit 0, got ${exit_code} (see ${log})"
  LAST_LOG="$log"
  pass "${name}: exit 0"
}

run_fail() {
  local name="$1"
  shift
  local log="${RESULT_DIR}/${name}.log"
  record_command_line "$@" >> "$TRANSCRIPT"
  local exit_code
  if "$@" > "$log" 2>&1; then
    exit_code=0
  else
    exit_code=$?
  fi
  cat "$log" >> "$TRANSCRIPT"
  print -r -- "[exit=${exit_code}]" >> "$TRANSCRIPT"
  print >> "$TRANSCRIPT"
  (( exit_code != 0 )) || fail "${name}: expected nonzero exit"
  LAST_LOG="$log"
  pass "${name}: failed closed with exit ${exit_code}"
}

assert_contains() {
  local file="$1"
  local literal="$2"
  local description="$3"
  if rg -F -q -- "$literal" "$file"; then
    pass "$description"
  else
    fail "$description: missing literal ${literal} in ${file}"
  fi
}

assert_not_contains() {
  local file="$1"
  local literal="$2"
  local description="$3"
  if rg -F -q -- "$literal" "$file"; then
    fail "$description: unexpected literal ${literal} in ${file}"
  else
    pass "$description"
  fi
}

assert_multiline_contains() {
  local file="$1"
  local literal="$2"
  local description="$3"
  if rg -U -F -q -- "$literal" "$file"; then
    pass "$description"
  else
    fail "$description: missing multiline literal in ${file}"
  fi
}

assert_literal_count() {
  local file="$1"
  local literal="$2"
  local expected="$3"
  local description="$4"
  local actual
  actual="$(rg -F -c -- "$literal" "$file" 2>/dev/null || true)"
  [[ -n "$actual" ]] || actual=0
  if [[ "$actual" == "$expected" ]]; then
    pass "$description (count=${actual})"
  else
    fail "$description: count=${actual}, expected=${expected} in ${file}"
  fi
}

assert_final_arg_count() {
  local file="$1"
  local argument="$2"
  local expected="$3"
  local description="$4"
  local section="${RESULT_DIR}/.codex-args-section.log"
  sed -n '/^codex_args:$/,$p' "$file" > "$section"
  assert_literal_count "$section" "  - \"${argument}\"" "$expected" "$description"
}

assert_files_equal() {
  local left="$1"
  local right="$2"
  local description="$3"
  if cmp -s "$left" "$right"; then
    pass "$description"
  else
    fail "$description: files differ (${left}, ${right})"
  fi
}

MODEL="primary-parent-${LABEL}"
CLI_MODEL="primary-cli-${LABEL}"
DUP_MODEL="primary-duplicate-${LABEL}"

printf '%s\n' \
  "# ${TASK_ID}: setup preservation fixture for ${LABEL}" \
  '[mcp]' \
  'enabled_servers = ["figma"]' \
  '' \
  '[custom.validation]' \
  "owner = \"${TASK_ID}\"" \
  "target = \"${LABEL}\"" \
  > "$ROOT_CONFIG"
cp "$ROOT_CONFIG" "${RESULT_DIR}/setup-preservation-before.toml"

print -r -- "fixture: ${ROOT_CONFIG}" >> "$TRANSCRIPT"
cat "$ROOT_CONFIG" >> "$TRANSCRIPT"
print >> "$TRANSCRIPT"

run_ok 01_setup_preserve \
  env "HOME=${HOME_DIR}" \
  "$BIN" setup local "$TARGET" --source-dir "$RUNTIME_SOURCE"
assert_files_equal "${RESULT_DIR}/setup-preservation-before.toml" "$ROOT_CONFIG" \
  'setup local without primary flags preserves project-config byte-for-byte'

run_ok 02_setup_set \
  env "HOME=${HOME_DIR}" \
  "$BIN" setup local "$TARGET" --source-dir "$RUNTIME_SOURCE" --no-sync \
  --codex-primary-model "$MODEL" \
  --codex-primary-reasoning-effort high \
  --codex-yolo-mode=true
assert_contains "$ROOT_CONFIG" '[mcp]' 'set preserves MCP table'
assert_contains "$ROOT_CONFIG" '[custom.validation]' 'set preserves unrelated table'
assert_contains "$ROOT_CONFIG" "model = '${MODEL}'" 'set writes model'
assert_contains "$ROOT_CONFIG" "reasoning_effort = 'high'" 'set writes reasoning effort'
assert_contains "$ROOT_CONFIG" 'yolo_mode = true' 'set writes boolean true'

run_ok 03_setup_update \
  env "HOME=${HOME_DIR}" \
  "$BIN" setup local "$TARGET" --source-dir "$RUNTIME_SOURCE" --no-sync \
  --codex-primary-reasoning-effort xhigh \
  --codex-yolo-mode=false
assert_contains "$ROOT_CONFIG" "model = '${MODEL}'" 'update preserves omitted model'
assert_contains "$ROOT_CONFIG" "reasoning_effort = 'xhigh'" 'update changes supplied reasoning effort'
assert_contains "$ROOT_CONFIG" 'yolo_mode = false' 'update preserves explicit false presence'

run_ok 04_setup_clear \
  env "HOME=${HOME_DIR}" \
  "$BIN" setup local "$TARGET" --source-dir "$RUNTIME_SOURCE" --no-sync \
  --clear-codex-primary-session
assert_not_contains "$ROOT_CONFIG" '[agents.codex.primary_session]' 'clear removes only primary-session table'
assert_contains "$ROOT_CONFIG" '[mcp]' 'clear preserves MCP table'
assert_contains "$ROOT_CONFIG" '[custom.validation]' 'clear preserves unrelated table'

run_ok 05_setup_reseed_parent \
  env "HOME=${HOME_DIR}" \
  "$BIN" setup local "$TARGET" --source-dir "$RUNTIME_SOURCE" --no-sync \
  --codex-primary-model "$MODEL" \
  --codex-primary-reasoning-effort high \
  --codex-yolo-mode=true

printf '%s\n' \
  "# ${TASK_ID}: child composition fixture for ${LABEL}" \
  '[mcp]' \
  'enabled_servers = ["safari"]' \
  '' \
  '[agents.codex.primary_session]' \
  'reasoning_effort = "medium"' \
  'yolo_mode = false' \
  > "$CHILD_CONFIG"
cp "$CHILD_CONFIG" "${RESULT_DIR}/valid-child-config.toml"

print -r -- "fixture: ${CHILD_CONFIG}" >> "$TRANSCRIPT"
cat "$CHILD_CONFIG" >> "$TRANSCRIPT"
print >> "$TRANSCRIPT"

run_ok 06_parent_print_config \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${TARGET}" \
  "$BIN" codex --print-config exec inspect
PARENT_PRINT="$LAST_LOG"
assert_contains "$PARENT_PRINT" "effective_value: \"${MODEL}\"" 'parent print-config applies project model'
assert_contains "$PARENT_PRINT" "effective_source: ${ROOT_CONFIG}" 'parent print-config records project provenance'
assert_contains "$PARENT_PRINT" 'effective_value: true' 'parent print-config applies yolo true'
assert_final_arg_count "$PARENT_PRINT" '--dangerously-bypass-approvals-and-sandbox' 1 \
  'project yolo true emits exactly one native danger flag'

run_ok 07_child_print_config \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config exec inspect
CHILD_PRINT="$LAST_LOG"
assert_contains "$CHILD_PRINT" "effective_value: \"${MODEL}\"" 'child inherits parent model'
assert_contains "$CHILD_PRINT" "project_source: ${ROOT_CONFIG}" 'nearest per-field provenance retains parent model source'
assert_contains "$CHILD_PRINT" 'effective_value: "medium"' 'child overrides parent reasoning effort'
assert_contains "$CHILD_PRINT" "project_source: ${CHILD_CONFIG}" 'nearest per-field provenance records child source'
assert_contains "$CHILD_PRINT" 'effective_value: false' 'child false masks inherited yolo true'
assert_contains "$CHILD_PRINT" '  - figma' 'MCP ordered union contains parent figma'
assert_contains "$CHILD_PRINT" '  - safari' 'MCP ordered union contains child safari'
assert_contains "$CHILD_PRINT" 'mcp_servers.figma.url=' 'final argv contains figma MCP override'
assert_contains "$CHILD_PRINT" 'mcp_servers.safari.command=' 'final argv contains safari MCP override'
assert_final_arg_count "$CHILD_PRINT" '--dangerously-bypass-approvals-and-sandbox' 0 \
  'child yolo false emits no danger flag'

run_ok 08_child_doctor \
  env "HOME=${HOME_DIR}" \
  "$BIN" doctor local "$CHILD"
CHILD_DOCTOR="$LAST_LOG"
assert_contains "$CHILD_DOCTOR" 'codex_primary_config_valid: true' 'doctor validates composed config'
assert_contains "$CHILD_DOCTOR" "codex_primary_model: ${MODEL}" 'doctor reports inherited model'
assert_contains "$CHILD_DOCTOR" "codex_primary_model_source: ${ROOT_CONFIG}" 'doctor reports model provenance'
assert_contains "$CHILD_DOCTOR" 'codex_primary_reasoning_effort: medium' 'doctor reports child effort'
assert_contains "$CHILD_DOCTOR" "codex_primary_reasoning_effort_source: ${CHILD_CONFIG}" 'doctor reports effort provenance'
assert_contains "$CHILD_DOCTOR" 'codex_primary_yolo_mode: false' 'doctor reports explicit false'
assert_contains "$CHILD_DOCTOR" "codex_primary_yolo_mode_source: ${CHILD_CONFIG}" 'doctor distinguishes false from default'
assert_contains "$CHILD_DOCTOR" 'codex_mcp_enabled: figma,safari' 'doctor reports MCP coexistence'

run_ok 09_parent_yolo_dedupe \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${TARGET}" \
  "$BIN" codex --print-config -d --danger --yolo \
  --dangerously-bypass-approvals-and-sandbox exec inspect
assert_final_arg_count "$LAST_LOG" '--dangerously-bypass-approvals-and-sandbox' 1 \
  'project plus duplicate explicit yolo requests normalize to one danger flag'

run_ok 10_child_explicit_yolo \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config --danger exec inspect
assert_contains "$LAST_LOG" 'effective_source: wrapper:--danger' 'explicit wrapper yolo source is visible'
assert_final_arg_count "$LAST_LOG" '--dangerously-bypass-approvals-and-sandbox' 1 \
  'explicit yolo overrides project false with exactly one danger flag'

run_ok 11_child_profile \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config --profile fast exec inspect
PROFILE_PRINT="$LAST_LOG"
assert_literal_count "$PROFILE_PRINT" 'project_application: suppressed_by_explicit_profile' 2 \
  'profile suppresses project model and reasoning independently'
assert_literal_count "$PROFILE_PRINT" 'effective_value: (codex-native)' 2 \
  'profile leaves both string dimensions Codex-native'
assert_final_arg_count "$PROFILE_PRINT" '--dangerously-bypass-approvals-and-sandbox' 0 \
  'profile does not turn child yolo false on'

run_ok 12_parent_profile_keeps_yolo \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${TARGET}" \
  "$BIN" codex --print-config --profile fast exec inspect
assert_literal_count "$LAST_LOG" 'project_application: suppressed_by_explicit_profile' 2 \
  'profile suppresses parent model and reasoning'
assert_final_arg_count "$LAST_LOG" '--dangerously-bypass-approvals-and-sandbox' 1 \
  'profile does not suppress project yolo true'

run_ok 13_explicit_cli_with_profile \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config \
  -c "model=\"${CLI_MODEL}\"" \
  -c 'model_reasoning_effort="low"' \
  --profile fast exec inspect
CLI_PRINT="$LAST_LOG"
assert_contains "$CLI_PRINT" "effective_value: \"${CLI_MODEL}\"" 'explicit top-level config model wins'
assert_contains "$CLI_PRINT" 'effective_value: "low"' 'explicit top-level config reasoning wins'
assert_literal_count "$CLI_PRINT" 'project_application: suppressed_by_explicit_cli' 2 \
  'explicit CLI dimensions take precedence even with profile'
assert_contains "$CLI_PRINT" '  - "--profile"' 'explicit profile passes through'
assert_final_arg_count "$CLI_PRINT" '--dangerously-bypass-approvals-and-sandbox' 0 \
  'explicit model/profile selection does not alter child yolo false'

run_ok 14_equal_duplicate_cli \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config \
  "--model=${DUP_MODEL}" "-m=${DUP_MODEL}" \
  "-c=model='${DUP_MODEL}'" \
  "--config=model_reasoning_effort='high'" \
  '-c=model_reasoning_effort="high"' \
  exec inspect
DUP_PRINT="$LAST_LOG"
assert_final_arg_count "$DUP_PRINT" "--model=${DUP_MODEL}" 1 \
  'equal duplicate model selections normalize to one override'
assert_final_arg_count "$DUP_PRINT" "-m=${DUP_MODEL}" 0 \
  'duplicate short model override is removed'
assert_final_arg_count "$DUP_PRINT" "-c=model='${DUP_MODEL}'" 0 \
  'duplicate top-level config model override is removed'
assert_final_arg_count "$DUP_PRINT" "--config=model_reasoning_effort='high'" 1 \
  'equal duplicate reasoning selections normalize to one override'
assert_final_arg_count "$DUP_PRINT" '-c=model_reasoning_effort="high"' 0 \
  'duplicate reasoning override is removed'

run_fail 15_conflicting_model_cli \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config --model first -m second
assert_contains "$LAST_LOG" 'conflicting explicit Codex values for field model' \
  'conflicting explicit model values fail before render/exec'

run_fail 16_conflicting_reasoning_cli \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config \
  -c 'model_reasoning_effort="high"' \
  --config 'model_reasoning_effort="medium"'
assert_contains "$LAST_LOG" 'conflicting explicit Codex values for field model_reasoning_effort' \
  'conflicting explicit reasoning values fail before render/exec'

cp "$CHILD_CONFIG" "${RESULT_DIR}/child-before-invalid.toml"
printf '%s\n' \
  '[mcp]' \
  'enabled_servers = ["safari"]' \
  '' \
  '[agents.codex.primary_session]' \
  'yolo_mode = "false"' \
  > "$CHILD_CONFIG"
cp "$CHILD_CONFIG" "${RESULT_DIR}/invalid-child-config-before-setup.toml"

run_fail 17_invalid_print_config \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "$BIN" codex --print-config exec invalid
assert_contains "$LAST_LOG" "$CHILD_CONFIG" 'invalid print-config error includes exact source path'
assert_contains "$LAST_LOG" 'agents.codex.primary_session.yolo_mode' 'invalid print-config error includes exact field'

run_fail 18_invalid_doctor \
  env "HOME=${HOME_DIR}" \
  "$BIN" doctor local "$CHILD"
assert_contains "$LAST_LOG" 'codex_primary_config_valid: false' 'invalid doctor reports stable validity field'
assert_contains "$LAST_LOG" "$CHILD_CONFIG" 'invalid doctor error includes exact source path'
assert_contains "$LAST_LOG" 'agents.codex.primary_session.yolo_mode' 'invalid doctor error includes exact field'
assert_not_contains "$LAST_LOG" 'codex_primary_model:' 'invalid doctor emits no partial primary values'

INVALID_SENTINEL="${RESULT_DIR}/invalid-launch-sentinel.log"
rm -f "$INVALID_SENTINEL"
run_fail 19_invalid_launch_pre_exec \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${CHILD}" \
  "PATH=${FAKE_BIN}:${PATH}" "FAKE_CODEX_SENTINEL=${INVALID_SENTINEL}" \
  "$BIN" codex exec invalid
if [[ -e "$INVALID_SENTINEL" ]]; then
  fail 'invalid launch invoked Codex instead of failing during config resolution'
else
  pass 'invalid launch fails before Codex exec'
fi

run_fail 20_invalid_setup_atomic \
  env "HOME=${HOME_DIR}" \
  "$BIN" setup local "$CHILD" --source-dir "$RUNTIME_SOURCE" --no-sync \
  --codex-primary-model should-not-write
assert_files_equal "${RESULT_DIR}/invalid-child-config-before-setup.toml" "$CHILD_CONFIG" \
  'invalid setup fails atomically without changing original bytes'

cp "${RESULT_DIR}/child-before-invalid.toml" "$CHILD_CONFIG"

cp "$ROOT_CONFIG" "${RESULT_DIR}/root-before-no-config.toml"
rm "$ROOT_CONFIG"

run_ok 21_no_config_print \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${TARGET}" \
  "$BIN" codex --print-config exec native
NO_CONFIG_PRINT="$LAST_LOG"
assert_multiline_contains "$NO_CONFIG_PRINT" $'project_configs:\n  - (none)\nprimary_session:' \
  'no-config print reports no discovered project configs'
assert_literal_count "$NO_CONFIG_PRINT" 'effective_value: (codex-native)' 2 \
  'no-config leaves model and reasoning Codex-native'
assert_contains "$NO_CONFIG_PRINT" 'effective_source: default' 'no-config yolo source is default'
assert_contains "$NO_CONFIG_PRINT" 'effective_value: false' 'no-config yolo is safely false'
assert_final_arg_count "$NO_CONFIG_PRINT" '--dangerously-bypass-approvals-and-sandbox' 0 \
  'no-config argv has no danger flag'
assert_final_arg_count "$NO_CONFIG_PRINT" '--model' 0 'no-config argv has no generated model override'
assert_not_contains "$NO_CONFIG_PRINT" 'model_reasoning_effort=' 'no-config argv has no generated reasoning override'
assert_not_contains "$NO_CONFIG_PRINT" 'task-board.config.json' 'primary launcher does not inspect task-board config'
assert_not_contains "$NO_CONFIG_PRINT" 'max_parallel' 'primary launcher does not consume task-board spawn policy'

POLICY_ONLY="${TARGET}/.validation/${TASK_ID}/task-board-policy-only"
POLICY_BOARD_CONFIG="${POLICY_ONLY}/task-board.config.json"
mkdir -p "$POLICY_ONLY"
printf '%s\n' \
  '{' \
  '  "mode": "local",' \
  '  "local": {"board_dir": ".task-board"},' \
  '  "spawn": {' \
  '    "max_parallel": 7,' \
  '    "ceilings": {' \
  '      "codex": {' \
  "        \"model\": \"task-board-only-${LABEL}\"," \
  '        "reasoning_effort": "max"' \
  '      }' \
  '    }' \
  '  }' \
  '}' \
  > "$POLICY_BOARD_CONFIG"
cp "$POLICY_BOARD_CONFIG" "${RESULT_DIR}/task-board-policy-only-before.json"

print -r -- "fixture: ${POLICY_BOARD_CONFIG}" >> "$TRANSCRIPT"
cat "$POLICY_BOARD_CONFIG" >> "$TRANSCRIPT"
print >> "$TRANSCRIPT"

run_ok 21b_task_board_policy_irrelevant \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${POLICY_ONLY}" \
  "$BIN" codex --print-config exec policy-only
POLICY_ONLY_PRINT="$LAST_LOG"
assert_multiline_contains "$POLICY_ONLY_PRINT" $'project_configs:\n  - (none)\nprimary_session:' \
  'spawn-ceiling-only project has no agents-infra project config'
assert_literal_count "$POLICY_ONLY_PRINT" 'effective_value: (codex-native)' 2 \
  'task-board spawn ceiling does not supply primary model or reasoning'
assert_not_contains "$POLICY_ONLY_PRINT" "task-board-only-${LABEL}" \
  'task-board ceiling model is absent from primary launch evidence'
assert_not_contains "$POLICY_ONLY_PRINT" 'task-board.config.json' \
  'task-board policy path is absent from primary launch evidence'
assert_final_arg_count "$POLICY_ONLY_PRINT" '--model' 0 \
  'task-board ceiling does not add a model flag'
assert_not_contains "$POLICY_ONLY_PRINT" 'model_reasoning_effort=' \
  'task-board ceiling does not add a reasoning override'
assert_final_arg_count "$POLICY_ONLY_PRINT" '--dangerously-bypass-approvals-and-sandbox' 0 \
  'task-board policy cannot enable primary yolo'
assert_files_equal "${RESULT_DIR}/task-board-policy-only-before.json" "$POLICY_BOARD_CONFIG" \
  'agents-infra leaves task-board spawn policy bytes unchanged'

run_ok 22_no_config_doctor \
  env "HOME=${HOME_DIR}" \
  "$BIN" doctor local "$TARGET"
NO_CONFIG_DOCTOR="$LAST_LOG"
assert_contains "$NO_CONFIG_DOCTOR" 'codex_primary_config_valid: true' 'no-config doctor remains valid'
assert_contains "$NO_CONFIG_DOCTOR" 'codex_primary_model: ' 'no-config doctor emits empty model value'
assert_contains "$NO_CONFIG_DOCTOR" 'codex_primary_model_source: native' 'no-config doctor reports native model source'
assert_contains "$NO_CONFIG_DOCTOR" 'codex_primary_reasoning_effort_source: native' 'no-config doctor reports native reasoning source'
assert_contains "$NO_CONFIG_DOCTOR" 'codex_primary_yolo_mode: false' 'no-config doctor reports safe false'
assert_contains "$NO_CONFIG_DOCTOR" 'codex_primary_yolo_mode_source: default' 'no-config doctor reports default yolo source'

NATIVE_SENTINEL="${RESULT_DIR}/native-launch-argv.log"
rm -f "$NATIVE_SENTINEL"
run_ok 23_no_config_native_launch \
  env "HOME=${HOME_DIR}" "AGENTS_INFRA_CALLER_CWD=${TARGET}" \
  "PATH=${FAKE_BIN}:${PATH}" "FAKE_CODEX_SENTINEL=${NATIVE_SENTINEL}" \
  "$BIN" codex exec native
printf '%s\n' exec native > "${RESULT_DIR}/expected-native-launch-argv.log"
assert_files_equal "${RESULT_DIR}/expected-native-launch-argv.log" "$NATIVE_SENTINEL" \
  'no-config launch passes native argv unchanged to Codex'

cp "${RESULT_DIR}/root-before-no-config.toml" "$ROOT_CONFIG"

assert_not_contains "$ROOT_CONFIG" 'spawn' 'project TOML contains no spawn policy'
assert_not_contains "$ROOT_CONFIG" 'ceiling' 'project TOML contains no spawn ceiling'
assert_not_contains "$CHILD_CONFIG" 'spawn' 'child project TOML contains no spawn policy'
assert_not_contains "$CHILD_CONFIG" 'ceiling' 'child project TOML contains no spawn ceiling'
assert_files_equal "$ORIGINAL_BOARD_CONFIG" "$BOARD_CONFIG" \
  'task-board.config.json remains byte-identical throughout validation'

cleanup
trap - EXIT
assert_files_equal "$ORIGINAL_ROOT_CONFIG" "$ROOT_CONFIG" \
  'disposable copy project config restored to its original bytes'
assert_files_equal "$ORIGINAL_BOARD_CONFIG" "$BOARD_CONFIG" \
  'task-board.config.json remains byte-identical after cleanup'

print -r -- "PASS ${LABEL}: primary-session validation matrix" > "${RESULT_DIR}/${TASK_ID}_${LABEL}-summary.txt"
print -r -- "PASS ${LABEL}: primary-session validation matrix"
