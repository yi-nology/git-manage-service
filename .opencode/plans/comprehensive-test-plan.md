# Comprehensive Test Plan

## Overview
Add ~33 test files with ~400-500 test cases covering the entire codebase.

## Layer 1: Pure Unit Tests (No External Dependencies)

### 1. `biz/utils/crypto_test.go`
- TestEncryptDecrypt_RoundTrip
- TestEncrypt_EmptyString
- TestDecrypt_EmptyString
- TestDecrypt_InvalidBase64
- TestDecrypt_CiphertextTooShort
- TestEncryptDecrypt_LongText
- TestEncryptDecrypt_SpecialCharacters (Chinese, emoji, JSON, HTML entities)
- TestEncrypt_DifferentCiphertexts (random IV verification)

### 2. `biz/service/notification/template_test.go`
- TestRenderTemplate_Success
- TestRenderTemplate_Empty
- TestRenderTemplate_InvalidSyntax
- TestRenderTemplate_CustomFuncs (default, truncate)
- TestValidateTemplate_Valid
- TestValidateTemplate_Invalid
- TestValidateTemplate_Empty
- TestRenderTitleAndContent_WithCustomTemplates
- TestRenderTitleAndContent_FallbackDefaults
- TestRenderTitleAndContent_NilData
- TestGetDefaultTitleTemplate_AllEvents
- TestGetDefaultContentTemplate_AllEvents
- TestGetAvailableVariables_NonEmpty
- TestFillDefaults_StatusText
- TestFillDefaults_TaskNameFallback

### 3. `biz/service/provider/detect_test.go`
- TestDetectPlatform_SSH_GitHub
- TestDetectPlatform_SSH_GitLab
- TestDetectPlatform_SSH_Gitea
- TestDetectPlatform_HTTPS_GitHub
- TestDetectPlatform_HTTP_GitLab
- TestDetectPlatform_SSHProtocol
- TestDetectPlatform_EmptyURL
- TestDetectPlatform_UnsupportedFormat
- TestDetectPlatform_InvalidSSHFormat
- TestDetectPlatform_InvalidHTTPPath
- TestDetectPlatform_StripGitSuffix
- TestClassifyHost_DefaultSelfHosted

### 4. `biz/service/lint/lint_service_test.go`
- TestHasField
- TestHasSection
- TestFindSectionLine
- TestApplyRule_RequiredName
- TestApplyRule_RequiredVersion
- TestApplyRule_RequiredRelease
- TestApplyRule_RequiredSummary
- TestApplyRule_RequiredLicense
- TestApplyRule_RequiredURL
- TestApplyRule_RequiredDescription
- TestApplyRule_RequiredPrep
- TestApplyRule_RequiredBuild
- TestApplyRule_RequiredInstall
- TestApplyRule_RequiredFiles
- TestApplyRule_EmptySections
- TestApplyRule_BuildRootUsage
- TestApplyRule_MacroConsistency
- TestApplyRule_ChangelogFormat
- TestApplyRule_NoTabs
- TestApplyRule_CustomRegex

### 5. `biz/service/lint/ai_lint_service_test.go`
- TestExtractJSON_ValidJSON
- TestExtractJSON_JSONInMarkdown
- TestExtractJSON_NoJSON
- TestExtractJSON_Empty
- TestParseAILintResponse_Valid
- TestParseAILintResponse_EmptyIssues
- TestParseAILintResponse_InvalidSeverity
- TestParseAILintResponse_NoJSON

### 6. `biz/service/spec/spec_service_validate_test.go`
- TestValidateSpec_MissingName
- TestValidateSpec_MissingVersion
- TestValidateSpec_AllRequired
- TestValidateSpec_ChangelogFormat
- TestValidateSpec_NoTabs
- TestValidateSpec_ValidSpec
- TestGetBuiltinRules_NonEmpty
- TestGetBuiltinRules_Count
- TestHasSection
- TestApplyRule_CustomRegex

### 7. `biz/service/stats/language_config_test.go`
- TestGetLanguageConfig_Go
- TestGetLanguageConfig_Python
- TestGetLanguageConfig_JavaScript
- TestGetLanguageConfig_Dockerfile
- TestGetLanguageConfig_Makefile
- TestGetLanguageConfig_Unknown
- TestGetLanguageConfig_CaseInsensitive
- TestGetSupportedExtensions_NonEmpty
- TestGetSupportedExtensions_ContainsCommon

### 8. `biz/middleware/cors_test.go`
- TestCORS_SetsHeaders
- TestCORS_OptionsReturns204
- TestCORS_OriginHeader

### 9. `biz/service/codereview/diff_parser_edge_test.go`
- TestParseDiff_BinaryFile
- TestParseDiff_RenameFile
- TestParseDiff_EmptyHunk
- TestParseDiff_MalformedDiff
- TestParseDiff_LargeFile
- TestMigrationRule_DestructiveDrop
- TestMigrationRule_ForeignKeyWithoutIndex
- TestMigrationRule_SkipDeletedFiles
- TestMigrationRule_NonMigrationFile

### 10. `biz/service/notification/sender_test.go`
- TestDingTalkSender_BuildPayload
- TestDingTalkSender_Sign
- TestFeishuSender_BuildPayload
- TestWebhookSender_BuildPayload

### 11. `biz/service/codereview/aggregator_test.go` (Extended)
- TestCalculateRisk_NoFindings
- TestCalculateRisk_CriticalFinding
- TestCalculateRisk_HighFinding
- TestCalculateRisk_MultipleMediumEscalates
- TestCalculateRisk_MultipleLowEscalates
- TestDeduplicate_Empty
- TestCountBySeverity
- TestAggregate_EmptyFindings

### 12. `biz/service/branchrule/branch_rule_service_test.go`
- TestBranchRuleToDTO
- TestDtoToRules
- TestGetDefaultRuleSetDTO

### 13. `biz/service/git/ssh_key_helper_test.go`
- TestDetectKeyType
- TestBuildSSHCommand

## Layer 2: DAO Tests (SQLite In-Memory)

### Infrastructure: `biz/dal/db/test_helper.go`
- SetupTestDB() function initializing SQLite in-memory with AutoMigrate

### 14-32. DAO Test Files (one per DAO)
Each with CRUD + edge case tests.

## Layer 3: Model Hook Tests

### 33. `biz/model/po/encrypt_hook_test.go`
- TestRepo_BeforeSaveAfterFind_Encryption
- TestCredential_BeforeSaveAfterFind_Encryption
- TestSSHKey_BeforeSaveAfterFind_Encryption
- TestLLMProvider_BeforeSaveAfterFind_Encryption

## Estimated Totals
- ~33 test files
- ~400-500 test cases
- Coverage focus on utils, service logic, DAO CRUD, model hooks
