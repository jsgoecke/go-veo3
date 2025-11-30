<!--
SYNC IMPACT REPORT
==================
Version Change: NEW → 1.2.0 (Initial Ratification)
Date: 2025-11-30

Principles Added (12 total):
  I.   Library-First with CLI Interface
  II.  Test-First (NON-NEGOTIABLE)
  III. Integration Testing
  IV.  Observability
  V.   Documentation First
  VI.  Quality Management System
  VII. APIs as First-Class Features
  VIII. Scope-Based Authorization
  IX.  Feature Flag Architecture
  X.   Backend/Frontend Isolation
  XI.  Built for Compliance
  XII. Semantic Versioning and Conventional Commits

Sections Added:
  - Development Workflow
  - Review Process
  - Governance (with subsections)

Templates Requiring Updates:
  ⚠ PENDING: .specify/templates/tasks-template.md
    - Line 11 states "Tests are OPTIONAL"
    - CONFLICTS with Principle II (Test-First NON-NEGOTIABLE)
    - Must be updated to make tests MANDATORY

  ⚠ PENDING: .specify/templates/plan-template.md
    - Line 34 placeholder "[Gates determined based on constitution file]"
    - Should reference actual principles for validation gates

  ✅ .specify/templates/spec-template.md
    - Already includes mandatory testing sections
    - Aligns with Test-First principle

Follow-up Actions Required:
  1. Update tasks-template.md to enforce mandatory testing
  2. Update plan-template.md Constitution Check section with actual gates
  3. Establish CI/CD pipeline with quality gates per Principle VI
  4. Configure coverage tooling for 80% threshold enforcement
-->

# vcon-pipeline Constitution

**Version**: 1.2.0  
**Ratified**: 2025-11-20  
**Last Amended**: 2025-11-25  
**Status**: Ratified

## Preamble

This constitution establishes the foundational principles and governance for the vcon-pipeline project. All development decisions, architectural choices, and implementation approaches must align with these principles. Adherence to this constitution ensures consistency, quality, and long-term maintainability of the codebase.

## Core Principles

### I. Library-First with CLI Interface

Complex business logic and reusable functionality should be extracted into standalone libraries where it makes sense. Libraries must be self-contained, independently testable, and thoroughly documented. Clear purpose required - no organizational-only libraries. Simple UI components or basic features can remain inline. Every library must expose functionality via CLI with text in/out protocol: stdin/args → stdout, errors → stderr. Support JSON + human-readable formats for consistent, scriptable interfaces.

**Rationale**: Library extraction promotes code reuse, enables independent testing, and facilitates integration with other systems. CLI interfaces ensure automation compatibility and debuggability.

### II. Test-First (NON-NEGOTIABLE)

**Tests ARE the implementation - not an optional add-on.** The only acceptable development sequence is:

1. **Write Tests First**: Define expected behavior through failing tests
2. **User Approval**: Confirm tests accurately represent requirements  
3. **Implement**: Write minimal code to make tests pass
4. **Refactor**: Improve design while maintaining passing tests

**MANDATORY REQUIREMENTS**:

- **Every feature MUST include comprehensive tests** covering:
  - Happy path (success cases)
  - Error paths (failure scenarios)
  - Edge cases (boundary conditions)
  - Integration points (if applicable)

- **Coverage Targets MUST Be Met**:
  - Minimum 80% line coverage for all new/modified code
  - 100% coverage of critical business logic
  - Coverage is measured automatically and enforced by CI/CD

- **Definition of "Done"**:
  - Feature is NOT done until tests exist and pass
  - Coverage targets are NOT optional - they must be met
  - No "implement now, test later" - tests come first, period

- **Explicit Prohibition**:
  - Writing implementation before tests is FORBIDDEN
  - Claiming feature completion without tests is FORBIDDEN
  - Requesting review of untested code is FORBIDDEN
  - Merging code that decreases coverage is FORBIDDEN

**ENFORCEMENT**: Any code submitted without tests will be rejected immediately. Any feature spec that doesn't include explicit test requirements will be sent back for revision. Any attempt to merge code that decreases coverage will be automatically blocked by CI/CD pipeline.

**Rationale**: Test-first development catches requirement misunderstandings early, provides executable specifications, and creates a safety net for refactoring. This is non-negotiable because untested code cannot be maintained with confidence. Testing is not a separate phase - it IS development.

### III. Integration Testing

Focus areas requiring integration tests: New library contract tests, Contract changes, Inter-service communication, Shared schemas. Integration tests must validate real-world usage scenarios and data flow between components.

**Rationale**: Unit tests alone cannot verify that components work correctly together. Integration tests catch interface mismatches, data transformation errors, and communication failures that only emerge when systems interact.

### IV. Observability

Text I/O ensures debuggability. Structured logging required for all operations. Performance metrics and error tracking must be implemented for production systems. All components must provide visibility into their internal state and operations.

**Rationale**: Production issues cannot be diagnosed without adequate observability. Structured logging and metrics enable rapid troubleshooting, performance optimization, and proactive monitoring.

### V. Documentation First

Comprehensive documentation must be created before implementation begins. All features require: specification documents, API documentation, user guides, and architectural decision records. Documentation must be maintained alongside code changes and validated for accuracy.

**Rationale**: Documentation-first approach clarifies requirements, prevents misunderstandings, and creates reference material that grows with the codebase. It ensures knowledge transfer and onboarding efficiency.

### VI. Quality Management System

**Code quality standards are non-negotiable and automatically enforced:**

**MANDATORY QUALITY GATES**:

- **Test Coverage**: 80% or better unit test coverage REQUIRED
  - Measured automatically on every commit
  - Pipeline BLOCKS merge if coverage drops below threshold
  - No exceptions without explicit written approval from architecture review board
  
- **Integration Tests**: All critical paths MUST have integration tests
  - Service boundaries
  - External API integrations  
  - Database operations
  - Message queue interactions

- **Code Quality**: All code MUST pass automated validation
  - Linting (zero warnings tolerated)
  - Static code analysis (security, complexity, maintainability)
  - Dependency vulnerability scanning
  - No "TODO" or "FIXME" in production code

**CI/CD ENFORCEMENT**:

- **Every commit triggers automated quality validation**
- **Pipeline MUST fail if any gate is not met**
- **No manual overrides allowed** - fix the code or fix the test
- **Coverage reports published on every PR** showing:
  - Current coverage percentage
  - Coverage delta (+ or -)
  - Uncovered lines highlighted
  - Comparison to main branch

**NO SHORTCUTS PERMITTED**:

- Cannot disable linting for convenience
- Cannot skip tests "just this once"
- Cannot merge code with failing tests
- Cannot commit untested code and "promise to add tests later"

**Rationale**: Quality gates prevent defects from reaching production, reduce technical debt, and maintain codebase health. Automated enforcement ensures consistent standards without relying on manual review. The 80% threshold is scientifically correlated with reduced production incidents and improved maintainability.

### VII. APIs as First-Class Features

All APIs are first-class features of the product with developer experience (DX) treated with the same importance as user experience (UX). Every API must be documented as if it were public-facing, regardless of initial intended use. No "private" APIs in terms of documentation quality or design standards. API design, usability, consistency, and documentation must meet the same standards as user-facing features. Decisions about what to actually expose publicly can be made later, but all APIs must be built to public standards from the start.

**Rationale**: Internal APIs often become external APIs as products evolve. Building to public standards from the start prevents costly refactoring and enables rapid feature expansion through third-party integrations.

### VIII. Scope-Based Authorization

All capabilities must be built with fine-grained scopes that enable precise authorization control. Every feature, API endpoint, and UI component must be associated with specific permission scopes. Authorization tokens carry scope assignments that determine accessible functionality. User interfaces must dynamically adapt based on available scopes - showing only features the current token can access. This enables flexible permission models, secure multi-tenant systems, and role-based access patterns without hardcoded authorization logic.

**Rationale**: Fine-grained scopes enable secure multi-tenancy, role-based access control, and principle of least privilege. Dynamic UI adaptation prevents security issues from confused authorization states.

### IX. Feature Flag Architecture

All new features must be implemented behind feature flags that enable controlled rollouts and safe deployment. Feature flags must support ring-based deployments for progressive testing, customer-specific feature assignments, and user-level targeting. This enables A/B testing, gradual rollouts, immediate rollback capabilities, and customized feature experiences. No feature should be permanently enabled without first being validated through feature flag controls across different user segments and deployment rings.

**Rationale**: Feature flags decouple deployment from release, enable risk mitigation through gradual rollouts, and provide immediate rollback capabilities without code changes. They support experimentation and customer-specific customization.

### X. Backend/Frontend Isolation

Clear architectural isolation must be maintained between backend and frontend systems. Frontend applications must never directly access databases or backend data stores - all data access must occur through well-defined APIs. This ensures that internal development patterns can seamlessly extend to external development when APIs are made publicly available. Backend services own data persistence and business logic, while frontend applications consume data exclusively through API contracts. This isolation enables third-party integrations, external development, and maintains proper separation of concerns.

**Rationale**: Backend/frontend isolation enforces API-first architecture, prevents data layer coupling, and enables independent scaling. It ensures APIs are truly usable by external developers because internal developers use the same interfaces.

### XI. Built for Compliance

All architectural, feature, and implementation decisions must take into account that all services and software developed will be audited for compliance. The compliance will include, but not be limited to, SOC2, ISO27001, HiTrust, HIPAA, etc. This ensures that decisions made will accelerate our ability to be audited for compliance and maintain that compliance. All decisions made for this should also be documented.

**Rationale**: Compliance requirements affect architecture, security controls, audit logging, data handling, and retention policies. Building with compliance in mind from the start prevents costly rearchitecture and ensures audit readiness.

### XII. Semantic Versioning and Conventional Commits

All project releases MUST follow Semantic Versioning 2.0.0 (semver.org) with version numbers in MAJOR.MINOR.PATCH format. Version increments follow strict rules: MAJOR for incompatible API changes, MINOR for backward-compatible functionality additions, PATCH for backward-compatible bug fixes. All commit messages MUST follow Conventional Commits 1.0.0 specification (conventionalcommits.org) using the format: `<type>(<scope>): <description>` with types including feat, fix, docs, style, refactor, test, chore. Breaking changes MUST include `BREAKING CHANGE:` footer or `!` after type/scope. This ensures clear communication of changes, enables automated changelog generation, and provides semantic meaning to version history.

**Rationale**: Semantic Versioning provides a universal language for communicating the impact of changes, enabling consumers to make informed upgrade decisions and maintain compatibility. Conventional Commits creates machine-readable history that enables automated tooling for changelog generation, version bumping, and release notes. Together, they form a robust system for version management that reduces human error, improves communication between maintainers and users, and enables sophisticated automation in CI/CD pipelines. This standardization is critical for compliance audits, dependency management, and maintaining trust with consumers of our APIs and services.

## Development Workflow

All development follows the specification-driven workflow:

1. **Requirements Gathering**: User requirements captured and validated
2. **Feature Specification**: Detailed spec created with user stories and acceptance criteria  
   - **MUST include explicit test requirements and success criteria**
3. **Implementation Plan**: Technical design, architecture decisions, and complexity analysis  
   - **MUST include test strategy and coverage targets**
4. **Task Breakdown**: Granular tasks organized by user story priority  
   - **MUST include dedicated test tasks (not optional)**
   - Test tasks MUST appear before implementation tasks
5. **Test-First Implementation**: **MANDATORY SEQUENCE**:
   - Write failing tests that define expected behavior
   - Get user approval that tests represent requirements correctly
   - Implement minimal code to make tests pass
   - Refactor while maintaining passing tests
   - Verify coverage targets are met
6. **Quality Validation**: Automated quality gates enforced by CI/CD  
   - **All gates MUST pass - no exceptions**
7. **Deployment**: Feature flag controlled rollout with observability

**ENFORCEMENT**: Each phase has defined deliverables and approval gates. No phase can be skipped. Test-first sequence is mandatory and automatically validated.

## Review Process

All code changes require appropriate review based on the scope of change:

- **Peer Review**: Required for all implementations to verify correctness and code quality  
  - **Reviewer MUST verify tests exist and pass**
  - **Reviewer MUST verify coverage targets are met**
  - **Reviewer MUST reject code without adequate tests**
  
- **Architecture Review**: Required for design changes affecting system structure or component interactions
- **Security Review**: Required for security-sensitive changes including authentication, authorization, data handling, and external integrations
- **Documentation Review**: Required for user-facing changes to ensure accuracy and completeness

All reviews must verify compliance with constitutional principles. Reviewers must explicitly confirm that changes align with relevant principles or document justified exceptions.

## Governance

### Amendment Process

1. **Proposal**: Amendment proposed with rationale and impact analysis
2. **Review Period**: Minimum 7 days for stakeholder feedback
3. **Approval**: Requires documented approval from project maintainers
4. **Migration Plan**: Implementation plan for bringing existing code into compliance
5. **Version Bump**: Constitution version incremented according to semantic versioning
6. **Announcement**: Changes communicated to all contributors

### Versioning Policy

Constitution versions follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Backward incompatible changes - principle removal, redefinition, or governance changes that invalidate existing practices
- **MINOR**: Backward compatible additions - new principles, expanded guidance, additional mandatory sections
- **PATCH**: Clarifications, wording improvements, typo fixes, non-semantic refinements

### Compliance Review

- Constitution supersedes all other practices and documentation
- All pull requests must include constitutional compliance verification
- Proposed complexity must be justified against simpler alternatives that align with principles
- Regular audits verify that codebase adheres to constitutional requirements
- Use runtime development guidance for implementation details not covered in the constitution

### Exception Handling

Exceptions to constitutional principles must be:

1. Explicitly documented with clear rationale
2. Reviewed and approved by project maintainers
3. Time-boxed with migration plan to compliance
4. Tracked as technical debt with remediation priority

Exceptions should be rare and only granted when benefits significantly outweigh costs.

### Enforcement Mechanisms

**Automated Enforcement**:
- CI/CD pipeline enforces all quality gates automatically
- Pre-commit hooks validate test file presence
- Coverage tools block merges that decrease coverage
- Static analysis prevents common anti-patterns

**Manual Enforcement**:
- Code reviews must verify constitutional compliance
- Architecture reviews validate design principles
- Regular audits identify compliance drift
- Technical debt tracked and prioritized for remediation

**Consequences for Non-Compliance**:
- Code without tests: **Immediate rejection, no review**
- Decreased coverage: **Automatic pipeline failure**  
- Skipped quality gates: **Merge blocked by automation**
- Repeated violations: **Escalation to architecture review board**

**The constitution is not optional guidance - it is mandatory policy with automated enforcement.**

---

*This constitution serves as the foundational governance document for the vcon-pipeline project. All contributors are expected to understand and follow these principles in their work. Testing is not optional. Quality is not negotiable. These standards exist to protect the project, our users, and our team.*
