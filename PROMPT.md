# Product Prompt

## Goal
Build a Docker Contailer that runs a Lumo API V2 and Go Service inside with a Web Front End that reads a user's Proton Mail and applies Gmail-like labels using Lumo to assign classifications.

## External Dependencies
- Proton Mail Go API: https://github.com/ProtonMail/go-proton-api
- Lumo API: https://github.com/carlostkd/Lumo-Api-V2

## V1 Product Requirements

### 1) User Interface and Authentication
- The user interface MUST be web based.
- Configuration MUST be stored in YAML.

### 2) Scan Behavior
- MUST run as a long-running command in daemon mode.
- Default scan interval MUST be every 5 minutes.
- Processing scope MUST be:
    - Inbox only.
    - Unread messages only.
- First-run behavior MUST be:
    - No historical mailbox backfill.
    - Process only new mail from the first successful login onward.

### 3) Lumo Classification Flow
- For each eligible email, it MUST send to Lumo:
    - Allowed labels list.
    - Email subject.
    - Email sender.
    - Redacted plain-text body.
- Body text MUST be derived from HTML-to-plain-text conversion and attachments MUST be ignored.
- Lumo response format in V1 MUST be plain text.
- Parsing strategy MUST be best-effort plain-text parsing against configured labels.
- Per-request Lumo timeout in V1 MUST be 20 seconds.
- Body text sent to Lumo MUST be capped at 2000 characters.

## Lumo Prompt
- MUST send Lumo this prompt template:

```text
Please reply with a label from the list of [<Proton Mail Labels>], for an email from [<Sender Email>] with subject [<Email Subject>] and the body [<Email Body Text>]
```

- It  MUST search Lumo output for configured Proton labels and select the matching label.
- If multiple configured labels are present in output, It SHOULD select the first label mention in the output text, unless Questionable is avalible.  If Questionable is in the list it MUST select Questionable.

### 4) Labeling Rules
- If Lumo returns a known label that exists in Proton, it MUST apply it.
- If Lumo returns a label that does not exist in Proton, it MUST skip it and log the event.
- Important handling:
    - If Lumo indicates Questionable, it MUST apply Proton label: Questionable
    - If Lumo indicates important, it MUST apply Proton label: Important, unless it is also Questionable.

### 5) Privacy and Safety
- Before sending body content to Lumo, it MUST redact sensitive patterns.
- Redaction patterns MUST be configurable via regex patterns in YAML config.
- Default redaction patterns in V1 MUST include emails, phone numbers, SSN, IBAN, and card numbers.
- Redaction output MUST use typed tokens (example: [REDACTED_EMAIL], [REDACTED_PHONE]).

### 6) Reliability and Errors
- Retry behavior for Proton and Lumo operations MUST be:
    - Retry up to 3 times on failure.
    - Use 30-second backoff between retries.
    - If retry fails, skip and continue.
- PMAIL MUST NOT crash daemon mode because a single message fails.
- PMAIL MUST log skips and failures in human-readable logs.
- Health check MUST report unhealthy for:
    - Proton unreachable.
    - Lumo unreachable.
    - Proton or Lumo authentication failure.
    - Daemon crash.
- Automatic repair MUST attempt retries and process restart on unhealthy state.
- If process-level recovery fails or unhealthy state persists, it MUST trigger a container restart.

### 7) State and Configuration
- It MUST use YAML files for both configuration and state.
- It MUST track processed message IDs so messages are not re-processed.
- It MUST persist the last scan checkpoint to support continuous polling.
- Processed IDs and decision logs MUST use a 30-day rolling retention policy.

### 8) Deployment, Runtime, and UX Decisions (Locked)
- Runtime model MUST be a single Docker container running multiple processes.
- Process supervision MUST use supervisord.
- Web stack MUST be React frontend + Go service.
- Web UI authentication MUST use local username/password with auto-generated admin credentials on first run.
- Proton/Lumo secret protection key MUST be provided via environment variable.
- V1 account support MUST be one Proton account.
- Label source MUST be configured allowlist intersected with Proton labels.
- Missing configured labels SHOULD be auto-created in Proton.
- Label precedence MUST be: if Questionable appears in Lumo output, apply Questionable even if Important also appears.
- Rate limits MUST be user-definable, with V1 defaults:
    - Max 10 emails per minute.
    - Max 20 emails per hour.
- Web UI default port MUST be 5866.
- Log output MUST go to console plus rotating files.
- Log rotation policy MUST be:
    - Max file size 16 MB.
    - Keep 8 rotated files.
- Persistent paths MUST be:
    - /lumo_lab/config
    - /lumo_lab/logs
    - /lumo_lab/state
- Timezone MUST be configurable, with default America/New_York.
- V1 UI scope MUST include:
    - Login/setup.
    - Config editor.
    - Label map preview.
    - Run status.
    - Recent decisions.
    - Logs.
    - Health view.


## Implementation Notes
- Keep architecture modular:
    - Proton client adapter.
    - Lumo client adapter.
    - Redaction module.
    - Processor/polling service.
    - Config and state stores.
- Prefer predictable behavior over aggressive automation.
- Optimize for safe defaults and transparent logs.

