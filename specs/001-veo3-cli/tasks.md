# Tasks: Veo3 CLI

**Input**: Design documents from `/specs/001-veo3-cli/`  
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Tests are MANDATORY per Constitution Principle II (Test-First NON-NEGOTIABLE). Every feature MUST include comprehensive tests with 80% minimum coverage.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project structure** (Go CLI)
- **cmd/veo3/** - Main entry point
- **pkg/** - Reusable library packages
- **internal/** - Private application code
- **tests/** - Integration tests and fixtures

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Initialize Go module with `go mod init github.com/yourorg/veo3-cli`
- [ ] T002 Create directory structure per plan.md (cmd/, pkg/, internal/, tests/)
- [ ] T003 [P] Create go.mod with initial dependencies (cobra, viper, google api client)
- [ ] T004 [P] Create .gitignore for Go projects (binaries, vendor/, coverage files)
- [ ] T005 [P] Create .golangci.yml with linter configuration (80% coverage threshold)
- [ ] T006 [P] Create Makefile with build, test, lint, and install targets
- [ ] T007 [P] Create README.md with installation and usage instructions
- [ ] T008 [P] Create main.go entry point in cmd/veo3/main.go with Cobra root command

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T009 [P] Create pkg/config/config.go with Configuration struct from data-model.md
- [ ] T010 [P] Create pkg/config/manager.go with config file load/save operations
- [ ] T011 [P] Create pkg/config/defaults.go with default configuration values
- [ ] T012 [P] Create internal/validation/files.go with image/video validation functions
- [ ] T013 [P] Create internal/validation/params.go with parameter validation functions
- [ ] T014 [P] Create pkg/veo3/client.go with Google API client wrapper and authentication
- [ ] T015 [P] Create pkg/veo3/models.go with model registry and capabilities from data-model.md
- [ ] T016 [P] Create pkg/operations/manager.go with operation lifecycle management
- [ ] T017 [P] Create pkg/operations/poller.go with status polling and exponential backoff
- [ ] T018 [P] Create pkg/operations/downloader.go with video download streaming
- [ ] T019 [P] Create internal/format/output.go with human-readable output formatting
- [ ] T020 [P] Create internal/format/json.go with JSON output formatting

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Text-to-Video Generation (Priority: P0) 🎯 MVP

**Goal**: Users can generate videos from text prompts with optional audio, duration, resolution, and aspect ratio options

**Independent Test**: Can be fully tested by executing `veo3 generate "test prompt"` and verifying a video file is created with expected properties

### Tests for User Story 1 (MANDATORY) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**
> **Tests are NON-NEGOTIABLE per Principle II - minimum 80% coverage required**

- [ ] T021 [P] [US1] Write unit tests for GenerationRequest validation in tests/unit/veo3/generate_test.go
- [ ] T022 [P] [US1] Write unit tests for prompt validation (length, content) in tests/unit/validation/params_test.go
- [ ] T023 [P] [US1] Write integration test for full generate command flow in tests/integration/cli_test.go
- [ ] T024 [P] [US1] Write API client mock tests for text-to-video request in tests/unit/veo3/client_test.go

### Implementation for User Story 1

- [ ] T025 [P] [US1] Create pkg/veo3/generate.go with GenerateVideo function and GenerationRequest struct
- [ ] T026 [P] [US1] Implement request validation in pkg/veo3/generate.go (prompt length, parameters)
- [ ] T027 [US1] Integrate Google API client call in pkg/veo3/generate.go (depends on T025, T026)
- [ ] T028 [P] [US1] Create pkg/cli/generate.go with cobra command for 'generate' subcommand
- [ ] T029 [US1] Wire up command flags (--model, --resolution, --duration, --aspect-ratio, --negative-prompt, --output)
- [ ] T030 [US1] Add progress display using progressbar library in pkg/cli/generate.go
- [ ] T031 [US1] Add operation polling and video download on completion in pkg/cli/generate.go
- [ ] T032 [US1] Add error handling with actionable messages in pkg/cli/generate.go
- [ ] T033 [US1] Add --json flag support for machine-readable output in pkg/cli/generate.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Image-to-Video Generation (Priority: P0)

**Goal**: Users can animate a static image into a video using an input image as the first frame

**Independent Test**: Can be tested by executing `veo3 animate test.png --prompt "animation"` and verifying video starts with provided image

### Tests for User Story 2 (MANDATORY) ⚠️

- [ ] T034 [P] [US2] Write unit tests for ImageRequest validation in tests/unit/veo3/animate_test.go
- [ ] T035 [P] [US2] Write unit tests for image file validation (format, size) in tests/unit/validation/files_test.go
- [ ] T036 [P] [US2] Write integration test for animate command in tests/integration/cli_test.go
- [ ] T037 [P] [US2] Write tests for base64 encoding of images in tests/unit/veo3/animate_test.go

### Implementation for User Story 2

- [ ] T038 [P] [US2] Create pkg/veo3/animate.go with AnimateImage function and ImageRequest struct
- [ ] T039 [P] [US2] Implement image file validation (magic bytes, size limit) in internal/validation/files.go
- [ ] T040 [US2] Implement base64 encoding for image upload in pkg/veo3/animate.go
- [ ] T041 [P] [US2] Create pkg/cli/animate.go with cobra command for 'animate' subcommand
- [ ] T042 [US2] Wire up command for image path positional arg and prompt flag
- [ ] T043 [US2] Add error handling for unsupported formats and size limits in pkg/cli/animate.go
- [ ] T044 [US2] Integrate with operation polling and download from US1

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Frame Interpolation (Priority: P1)

**Goal**: Users can generate videos by specifying both first and last frames with smooth interpolation

**Independent Test**: Can be tested by executing `veo3 interpolate start.png end.png` and verifying smooth transition between frames

### Tests for User Story 3 (MANDATORY) ⚠️

- [ ] T045 [P] [US3] Write unit tests for InterpolationRequest validation in tests/unit/veo3/interpolate_test.go
- [ ] T046 [P] [US3] Write unit tests for compatible dimensions check in tests/unit/validation/files_test.go
- [ ] T047 [P] [US3] Write integration test for interpolate command in tests/integration/cli_test.go

### Implementation for User Story 3

- [ ] T048 [P] [US3] Create pkg/veo3/interpolate.go with InterpolateFrames function and InterpolationRequest struct
- [ ] T049 [US3] Implement dual image loading and validation in pkg/veo3/interpolate.go
- [ ] T050 [US3] Enforce constraints (8s duration, 16:9 aspect) in pkg/veo3/interpolate.go
- [ ] T051 [P] [US3] Create pkg/cli/interpolate.go with cobra command for 'interpolate' subcommand
- [ ] T052 [US3] Wire up command for two positional args (first and last frame paths)
- [ ] T053 [US3] Add compatibility validation error messages in pkg/cli/interpolate.go

**Checkpoint**: All P0 and first P1 story should now be independently functional

---

## Phase 6: User Story 4 - Reference Image-Guided Generation (Priority: P1)

**Goal**: Users can provide up to 3 reference images to guide content, style, and subject appearance

**Independent Test**: Can be tested by executing `veo3 generate "prompt" --reference img1.png --reference img2.png` and verifying visual consistency

### Tests for User Story 4 (MANDATORY) ⚠️

- [ ] T054 [P] [US4] Write unit tests for ReferenceImageRequest validation in tests/unit/veo3/reference_test.go
- [ ] T055 [P] [US4] Write unit tests for 1-3 reference image count validation in tests/unit/veo3/reference_test.go
- [ ] T056 [P] [US4] Write integration test for reference images in tests/integration/cli_test.go

### Implementation for User Story 4

- [ ] T057 [P] [US4] Create pkg/veo3/reference.go with reference image handling and ReferenceImageRequest struct
- [ ] T058 [US4] Implement multi-file validation and base64 encoding in pkg/veo3/reference.go
- [ ] T059 [US4] Enforce model constraints (Veo 3.1 only, 8s, 16:9) in pkg/veo3/reference.go
- [ ] T060 [US4] Add --reference flag support to pkg/cli/generate.go (repeatable flag)
- [ ] T061 [US4] Add validation for max 3 references with clear error in pkg/cli/generate.go

**Checkpoint**: Reference-guided generation ready for brand consistency workflows

---

## Phase 7: User Story 5 - Video Extension (Priority: P1)

**Goal**: Users can extend Veo-generated videos by up to 7 seconds, chainable for longer content

**Independent Test**: Can be tested by generating a video, then extending it with `veo3 extend video.mp4 --prompt "continuation"`

### Tests for User Story 5 (MANDATORY) ⚠️

- [ ] T062 [P] [US5] Write unit tests for ExtensionRequest validation in tests/unit/veo3/extend_test.go
- [ ] T063 [P] [US5] Write unit tests for video validation (duration, format) in tests/unit/validation/files_test.go
- [ ] T064 [P] [US5] Write integration test for extend command in tests/integration/cli_test.go

### Implementation for User Story 5

- [ ] T065 [P] [US5] Create pkg/veo3/extend.go with ExtendVideo function and ExtensionRequest struct
- [ ] T066 [US5] Implement video file validation (Veo-generated, max 141s) in internal/validation/files.go
- [ ] T067 [US5] Implement video base64 encoding for upload in pkg/veo3/extend.go
- [ ] T068 [P] [US5] Create pkg/cli/extend.go with cobra command for 'extend' subcommand
- [ ] T069 [US5] Wire up command for video path positional arg and extension prompt flag
- [ ] T070 [US5] Add validation error messages for non-Veo videos in pkg/cli/extend.go

**Checkpoint**: All P1 stories complete - core generation features fully implemented

---

## Phase 8: User Story 6 - Operation Management (Priority: P2)

**Goal**: Users can list, check status, download, and cancel long-running operations

**Independent Test**: Can be tested by starting async generation, then using operations commands to manage it

### Tests for User Story 6 (MANDATORY) ⚠️

- [ ] T071 [P] [US6] Write unit tests for operation listing in tests/unit/operations/manager_test.go
- [ ] T072 [P] [US6] Write unit tests for status checking in tests/unit/operations/manager_test.go
- [ ] T073 [P] [US6] Write unit tests for download and cancel operations in tests/unit/operations/manager_test.go
- [ ] T074 [P] [US6] Write integration tests for operations subcommands in tests/integration/cli_test.go

### Implementation for User Story 6

- [ ] T075 [US6] Enhance pkg/operations/manager.go with list, status, download, cancel methods
- [ ] T076 [P] [US6] Create pkg/cli/operations.go with cobra command group for 'operations'
- [ ] T077 [P] [US6] Implement 'operations list' subcommand in pkg/cli/operations.go
- [ ] T078 [P] [US6] Implement 'operations status <id>' subcommand in pkg/cli/operations.go
- [ ] T079 [P] [US6] Implement 'operations download <id>' subcommand in pkg/cli/operations.go
- [ ] T080 [P] [US6] Implement 'operations cancel <id>' subcommand in pkg/cli/operations.go
- [ ] T081 [US6] Add --async flag to generate/animate/interpolate/extend commands
- [ ] T082 [US6] Add operation ID display for async mode in all generation commands

**Checkpoint**: Full async operation lifecycle management available

---

## Phase 9: User Story 7 - Model Selection and Information (Priority: P2)

**Goal**: Users can list available models, view capabilities, and select specific versions

**Independent Test**: Can be tested by executing `veo3 models list` and `veo3 models info <model>`

### Tests for User Story 7 (MANDATORY) ⚠️

- [ ] T083 [P] [US7] Write unit tests for model registry in tests/unit/veo3/models_test.go
- [ ] T084 [P] [US7] Write unit tests for model validation logic in tests/unit/veo3/models_test.go
- [ ] T085 [P] [US7] Write integration tests for models commands in tests/integration/cli_test.go

### Implementation for User Story 7

- [ ] T086 [US7] Enhance pkg/veo3/models.go with complete model registry from spec.md
- [ ] T087 [US7] Implement model capabilities and constraints checks in pkg/veo3/models.go
- [ ] T088 [P] [US7] Create pkg/cli/models.go with cobra command group for 'models'
- [ ] T089 [P] [US7] Implement 'models list' subcommand with formatted table output in pkg/cli/models.go
- [ ] T090 [P] [US7] Implement 'models info <model>' subcommand with detailed specs in pkg/cli/models.go
- [ ] T091 [US7] Add default model configuration support in pkg/config/config.go

**Checkpoint**: Model discovery and selection fully functional

---

## Phase 10: User Story 8 - Configuration Management (Priority: P2)

**Goal**: Users can configure API credentials, defaults, and preferences

**Independent Test**: Can be tested by running `veo3 config init`, setting values, and verifying commands use them

### Tests for User Story 8 (MANDATORY) ⚠️

- [ ] T092 [P] [US8] Write unit tests for configuration loading in tests/unit/config/manager_test.go
- [ ] T093 [P] [US8] Write unit tests for configuration saving in tests/unit/config/manager_test.go
- [ ] T094 [P] [US8] Write integration tests for config commands in tests/integration/cli_test.go
- [ ] T095 [P] [US8] Write tests for configuration precedence (flag > env > file) in tests/integration/config_test.go

### Implementation for User Story 8

- [ ] T096 [US8] Enhance pkg/config/manager.go with interactive init, get, set, show, reset methods
- [ ] T097 [US8] Implement secure file permissions (0600) for config file in pkg/config/manager.go
- [ ] T098 [US8] Implement XDG Base Directory specification support in pkg/config/manager.go
- [ ] T099 [P] [US8] Create pkg/cli/config.go with cobra command group for 'config'
- [ ] T100 [P] [US8] Implement 'config init' interactive setup in pkg/cli/config.go
- [ ] T101 [P] [US8] Implement 'config set <key> <value>' subcommand in pkg/cli/config.go
- [ ] T102 [P] [US8] Implement 'config show' with masked sensitive data in pkg/cli/config.go
- [ ] T103 [P] [US8] Implement 'config reset' subcommand in pkg/cli/config.go
- [ ] T104 [US8] Wire configuration loading into all commands via Viper

**Checkpoint**: Full configuration system operational

---

## Phase 11: User Story 9 - Batch Processing (Priority: P3)

**Goal**: Users can process multiple generation requests from a YAML manifest file

**Independent Test**: Can be tested by creating a manifest with multiple jobs and executing `veo3 batch process manifest.yaml`

### Tests for User Story 9 (MANDATORY) ⚠️

- [ ] T105 [P] [US9] Write unit tests for manifest parsing in tests/unit/batch/manifest_test.go
- [ ] T106 [P] [US9] Write unit tests for batch processor in tests/unit/batch/processor_test.go
- [ ] T107 [P] [US9] Write integration tests for batch commands in tests/integration/cli_test.go

### Implementation for User Story 9

- [ ] T108 [P] [US9] Create pkg/batch/manifest.go with YAML manifest parsing and BatchManifest struct
- [ ] T109 [US9] Implement job validation and type routing in pkg/batch/manifest.go
- [ ] T110 [P] [US9] Create pkg/batch/processor.go with concurrent job execution (worker pool pattern)
- [ ] T111 [US9] Implement progress tracking across all jobs in pkg/batch/processor.go
- [ ] T112 [US9] Implement summary report generation in pkg/batch/processor.go
- [ ] T113 [P] [US9] Create pkg/cli/batch.go with cobra command group for 'batch'
- [ ] T114 [P] [US9] Implement 'batch process <manifest>' subcommand in pkg/cli/batch.go
- [ ] T115 [P] [US9] Implement 'batch template' to generate sample manifest in pkg/cli/batch.go
- [ ] T116 [P] [US9] Implement 'batch retry <results.json>' for failed jobs in pkg/cli/batch.go
- [ ] T117 [US9] Add --concurrency flag for parallel job control in pkg/cli/batch.go

**Checkpoint**: Bulk generation automation ready for production workflows

---

## Phase 12: User Story 10 - Prompt Templates (Priority: P3)

**Goal**: Users can save and reuse prompt templates with variable substitution

**Independent Test**: Can be tested by saving a template, then generating with it using variable substitution

### Tests for User Story 10 (MANDATORY) ⚠️

- [ ] T118 [P] [US10] Write unit tests for template parsing in tests/unit/templates/parser_test.go
- [ ] T119 [P] [US10] Write unit tests for variable substitution in tests/unit/templates/parser_test.go
- [ ] T120 [P] [US10] Write integration tests for templates commands in tests/integration/cli_test.go

### Implementation for User Story 10

- [ ] T121 [P] [US10] Create pkg/templates/manager.go with template storage and retrieval (YAML file)
- [ ] T122 [P] [US10] Create pkg/templates/parser.go with Mustache-style {{variable}} substitution
- [ ] T123 [US10] Implement variable extraction from template strings in pkg/templates/parser.go
- [ ] T124 [P] [US10] Create pkg/cli/templates.go with cobra command group for 'templates'
- [ ] T125 [P] [US10] Implement 'templates save <name> --prompt <template>' in pkg/cli/templates.go
- [ ] T126 [P] [US10] Implement 'templates list' subcommand in pkg/cli/templates.go
- [ ] T127 [P] [US10] Implement 'templates export/import' for sharing in pkg/cli/templates.go
- [ ] T128 [US10] Add --template flag to generate command with --var support in pkg/cli/generate.go

**Checkpoint**: All user stories complete - full feature set implemented

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T129 [P] Create comprehensive README.md with examples and troubleshooting
- [ ] T130 [P] Generate man pages from Cobra documentation
- [ ] T131 [P] Add shell completion generation (bash, zsh, fish)
- [ ] T132 [P] Add --version command with build info (version, commit, build date)
- [ ] T133 [P] Implement structured logging with --verbose flag across all commands
- [ ] T134 Code cleanup and consistent error messages across all commands
- [ ] T135 Performance profiling and optimization for large batch jobs
- [ ] T136 [P] Security audit (API key handling, file permissions, input validation)
- [ ] T137 Add quickstart.md validation by running examples from it
- [ ] T138 [P] Create CONTRIBUTING.md with development setup instructions
- [ ] T139 [P] Create LICENSE file (Apache 2.0 or MIT)
- [ ] T140 Final coverage report and ensure 80% threshold is met

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-12)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P0 → P1 → P2 → P3)
- **Polish (Phase N)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P0)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P0)**: Can start after Foundational - Reuses polling/download from US1
- **User Story 3 (P1)**: Can start after Foundational - Reuses image validation from US2
- **User Story 4 (P1)**: Can start after Foundational - Reuses image handling from US2
- **User Story 5 (P1)**: Can start after Foundational - Reuses operation management
- **User Story 6 (P2)**: Can start after Foundational - Enhances operation system used by US1-5
- **User Story 7 (P2)**: Can start after Foundational - Enhances model selection from US1
- **User Story 8 (P2)**: Can start after Foundational - Enhances configuration from Foundation
- **User Story 9 (P3)**: Can start after Foundational - Orchestrates US1-5 functionality
- **User Story 10 (P3)**: Can start after Foundational - Enhances US1 with templates

### Within Each User Story

- Tests MUST be written and FAIL before implementation (Constitution Principle II)
- Models/structs before functions
- Core logic before CLI commands
- Validation before API calls
- Error handling integrated throughout
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models/files within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Implementation Strategy

### MVP First (User Stories 1-2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Text-to-Video)
4. Complete Phase 4: User Story 2 (Image-to-Video)
5. **STOP and VALIDATE**: Test both stories independently
6. Deploy/demo if ready

**MVP Deliverable**: CLI that can generate videos from text or images with full operation management

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Stories 3-5 (P1) → Test independently → Deploy/Demo
5. Add User Stories 6-8 (P2) → Test independently → Deploy/Demo
6. Add User Stories 9-10 (P3) → Test independently → Deploy/Demo
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 + User Story 6 (operations)
   - Developer B: User Story 2 + User Story 7 (models)
   - Developer C: User Story 3 + User Story 8 (config)
   - Developer D: User Stories 4-5 (reference + extension)
   - Developer E: User Stories 9-10 (batch + templates)
3. Stories complete and integrate independently

---

## Coverage & Quality Requirements

### Test Coverage (Constitution Principle II & VI)

**MANDATORY TARGETS**:
- Overall coverage: ≥80% (measured by `go test -coverprofile`)
- Critical packages: 100% coverage required
  - pkg/veo3/ (API client and generation logic)
  - pkg/config/ (configuration management)
  - pkg/operations/ (operation lifecycle)
- All exported functions must have tests
- All error paths must be tested

**Coverage Measurement**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go tool cover -func=coverage.out | grep total
```

**CI/CD Enforcement**:
- Pipeline BLOCKS merge if coverage < 80%
- Coverage reports published on every PR
- Coverage delta must be ≥0 (no decreases allowed)

### Quality Gates

**Linting** (golangci-lint):
- Zero warnings tolerated
- Runs on: `make lint`
- Configured in: `.golangci.yml`

**Static Analysis**:
- Security scanning: gosec
- Code complexity: gocyclo (max complexity: 15)
- Code formatting: gofmt, goimports

**Integration Tests**:
- All CLI commands must have end-to-end tests
- API interactions tested with mocked responses
- Configuration file persistence tested

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Tests are MANDATORY** per Constitution - every feature needs comprehensive tests
- 80% coverage threshold is NON-NEGOTIABLE
- Use `make test` to run all tests with coverage
- Use `make lint` to verify code quality before commits

---

## Task Count Summary

- **Setup**: 8 tasks
- **Foundational**: 12 tasks
- **User Story 1 (P0)**: 13 tasks (4 test + 9 implementation)
- **User Story 2 (P0)**: 11 tasks (4 test + 7 implementation)
- **User Story 3 (P1)**: 9 tasks (3 test + 6 implementation)
- **User Story 4 (P1)**: 8 tasks (3 test + 5 implementation)
- **User Story 5 (P1)**: 9 tasks (3 test + 6 implementation)
- **User Story 6 (P2)**: 12 tasks (4 test + 8 implementation)
- **User Story 7 (P2)**: 9 tasks (3 test + 6 implementation)
- **User Story 8 (P2)**: 13 tasks (4 test + 9 implementation)
- **User Story 9 (P3)**: 13 tasks (3 test + 10 implementation)
- **User Story 10 (P3)**: 11 tasks (3 test + 8 implementation)
- **Polish**: 12 tasks

**Total**: 140 tasks (34 test tasks + 106 implementation tasks)

**Parallel Opportunities**: 67 tasks marked [P] can run concurrently

**Test Coverage**: 34 dedicated test tasks ensuring 80% coverage minimum