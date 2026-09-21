# ============================================================
# ClipSense GitHub Project Kanban Synchronization Script
# ============================================================
# Run from inside the ClipSense repository.
#
# Purpose:
# - Reconfigures the default GitHub Project Status field into the
#   approved ClipSense Kanban workflow.
# - Reuses the 57 existing refined GitHub issues by exact title.
# - Adds missing issues to the project without recreating issues.
# - Places every issue in its intended Kanban column.
# - Synchronizes Priority, Type, and Scope project fields.
# - Synchronizes type, priority, scope, and needs-refinement labels.
#
# Final Kanban columns, in order:
# 1. Product Backlog
# 2. Needs Refinement
# 3. Ready
# 4. In Progress
# 5. QA / In Review
# 6. Blocked
# 7. Done
# 8. Icebox
#
# Safe behavior:
# - Does not create, delete, close, reopen, or rewrite issues.
# - Does not change issue bodies.
# - Does not modify product source code.
# - Does not move an invalid issue into Ready.
# - Supports a no-write DryRun.
# ============================================================

[CmdletBinding()]
param(
    [switch]$DryRun,
    [switch]$SelfTest
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$ProjectTitle = "Clipsense"
$ProjectLimit = 1000

$DesiredStatusOptions = @(
    [pscustomobject]@{ name = "Product Backlog";  color = "GRAY";   description = "Valid work that is not yet ready to start" },
    [pscustomobject]@{ name = "Needs Refinement"; color = "YELLOW"; description = "Work requiring decisions, evidence, scope, or acceptance criteria" },
    [pscustomobject]@{ name = "Ready";            color = "BLUE";   description = "Refined and independently actionable work" },
    [pscustomobject]@{ name = "In Progress";      color = "YELLOW"; description = "Work actively being implemented" },
    [pscustomobject]@{ name = "QA / In Review";   color = "PURPLE"; description = "Code review, functional validation, and acceptance" },
    [pscustomobject]@{ name = "Blocked";          color = "RED";    description = "Started work that cannot proceed because of a verified blocker" },
    [pscustomobject]@{ name = "Done";             color = "GREEN";  description = "Implemented, tested, reviewed, and accepted" },
    [pscustomobject]@{ name = "Icebox";           color = "GRAY";   description = "Accepted future work outside the active MVP delivery path" }
)

$DesiredPriorityOptions = @(
    [pscustomobject]@{ name = "P0"; color = "RED";    description = "Blocks the core ZIP MVP workflow or its acceptance" },
    [pscustomobject]@{ name = "P1"; color = "ORANGE"; description = "Required for a credible and usable MVP" },
    [pscustomobject]@{ name = "P2"; color = "YELLOW"; description = "Important stabilization or roadmap work" },
    [pscustomobject]@{ name = "P3"; color = "GRAY";   description = "Post-MVP or optional future work" }
)

$DesiredTypeOptions = @(
    [pscustomobject]@{ name = "User Story";         color = "PURPLE"; description = "User or stakeholder outcome" },
    [pscustomobject]@{ name = "Technical Enabler";  color = "GREEN";  description = "Engineering work that enables product delivery" },
    [pscustomobject]@{ name = "Bug";                color = "RED";    description = "Broken, failing, miswired, or defective behavior" },
    [pscustomobject]@{ name = "Security";           color = "RED";    description = "Security, dependency risk, validation, or vulnerability work" },
    [pscustomobject]@{ name = "QA";                 color = "BLUE";   description = "Testing, verification, smoke tests, and regression checks" },
    [pscustomobject]@{ name = "Documentation";      color = "BLUE";   description = "Documentation and repository-truth updates" },
    [pscustomobject]@{ name = "Chore";              color = "GRAY";   description = "Maintenance, setup, tooling, or operational work" }
)

$DesiredScopeOptions = @(
    [pscustomobject]@{ name = "MVP";           color = "BLUE";   description = "Required for the stabilized ZIP MVP" },
    [pscustomobject]@{ name = "Stabilization"; color = "PURPLE"; description = "MVP hardening, reliability, and cleanup" },
    [pscustomobject]@{ name = "Roadmap";       color = "YELLOW"; description = "Broader planned product work" },
    [pscustomobject]@{ name = "Post-MVP";      color = "GRAY";   description = "Future expansion outside the current MVP" }
)

$TypeLabelToFieldValue = @{
    "type:user-story"        = "User Story"
    "type:technical-enabler" = "Technical Enabler"
    "type:bug"               = "Bug"
    "type:security"          = "Security"
    "type:qa"                = "QA"
    "type:doc"               = "Documentation"
    "type:chore"             = "Chore"
}

$ScopeLabelToFieldValue = @{
    "scope:mvp"           = "MVP"
    "scope:stabilization" = "Stabilization"
    "scope:roadmap"       = "Roadmap"
    "scope:post-mvp"      = "Post-MVP"
}

$LabelsToEnsure = @(
    @{ Name = "type:user-story";        Color = "5319e7"; Description = "User or stakeholder outcome" },
    @{ Name = "type:technical-enabler"; Color = "0e8a16"; Description = "Engineering work that enables product delivery" },
    @{ Name = "type:bug";               Color = "d73a4a"; Description = "Broken, failing, miswired, or defective behavior" },
    @{ Name = "type:qa";                Color = "1d76db"; Description = "Testing, verification, smoke tests, and regression checks" },
    @{ Name = "type:security";          Color = "b60205"; Description = "Security, dependency risk, validation, or vulnerability work" },
    @{ Name = "type:doc";               Color = "0075ca"; Description = "Documentation and repository-truth updates" },
    @{ Name = "type:chore";             Color = "cfd3d7"; Description = "Maintenance, setup, tooling, or operational work" },
    @{ Name = "needs-refinement";        Color = "d876e3"; Description = "Issue requires more refinement before it is ready to implement" },

    @{ Name = "priority:p0"; Color = "b60205"; Description = "Critical blocker or MVP-breaking work" },
    @{ Name = "priority:p1"; Color = "d93f0b"; Description = "High-priority MVP or stabilization work" },
    @{ Name = "priority:p2"; Color = "fbca04"; Description = "Important but not immediately blocking" },
    @{ Name = "priority:p3"; Color = "c2e0c6"; Description = "Later, roadmap, or post-MVP work" },

    @{ Name = "scope:mvp";           Color = "0052cc"; Description = "Required for the stabilized ZIP MVP" },
    @{ Name = "scope:stabilization"; Color = "5319e7"; Description = "MVP stabilization and hardening" },
    @{ Name = "scope:roadmap";       Color = "c2e0c6"; Description = "Planned broader product work" },
    @{ Name = "scope:post-mvp";      Color = "eeeeee"; Description = "Later product expansion outside the current MVP" }
)

function Write-Section {
    param([Parameter(Mandatory)][string]$Title)

    Write-Host ""
    Write-Host "=== $Title ===" -ForegroundColor Cyan
}

function Invoke-GhJson {
    param([Parameter(Mandatory)][string[]]$Arguments)

    $Output = & gh @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "GitHub CLI command failed: gh $($Arguments -join ' ')"
    }

    $Text = ($Output | Out-String).Trim()
    if ([string]::IsNullOrWhiteSpace($Text)) {
        return $null
    }

    return $Text | ConvertFrom-Json
}

function Invoke-GhCommand {
    param([Parameter(Mandatory)][string[]]$Arguments)

    & gh @Arguments | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "GitHub CLI command failed: gh $($Arguments -join ' ')"
    }
}

function Test-BodyContains {
    param(
        [AllowNull()][string]$Body,
        [Parameter(Mandatory)][string]$Needle
    )

    if ($null -eq $Body) {
        return $false
    }

    return $Body.IndexOf($Needle, [System.StringComparison]::OrdinalIgnoreCase) -ge 0
}

function Get-CardTypeLabel {
    param([Parameter(Mandatory)][string[]]$Labels)

    $TypeLabels = @($Labels | Where-Object { $_ -like "type:*" })
    if ($TypeLabels.Count -ne 1) {
        return $null
    }

    if (-not $TypeLabelToFieldValue.ContainsKey($TypeLabels[0])) {
        return $null
    }

    return $TypeLabels[0]
}

function Get-CardScopeLabel {
    param([Parameter(Mandatory)][string[]]$Labels)

    $ScopeLabels = @($Labels | Where-Object { $_ -like "scope:*" })
    if ($ScopeLabels.Count -ne 1) {
        return $null
    }

    if (-not $ScopeLabelToFieldValue.ContainsKey($ScopeLabels[0])) {
        return $null
    }

    return $ScopeLabels[0]
}

function Test-StructuredIssueBody {
    param(
        [AllowNull()][string]$Body,
        [Parameter(Mandatory)][string[]]$Labels
    )

    if ([string]::IsNullOrWhiteSpace($Body)) {
        return $false
    }

    foreach ($Marker in @(
        "## Kanban Placement",
        "## Agile Rule",
        "Repository evidence can inform this card"
    )) {
        if (Test-BodyContains -Body $Body -Needle $Marker) {
            return $false
        }
    }

    $CardType = Get-CardTypeLabel -Labels $Labels
    if (-not $CardType) {
        return $false
    }

    $RequiredSectionsByType = @{
        "type:user-story"        = @("## User Story", "Acceptance Criteria")
        "type:bug"               = @("Acceptance Criteria")
        "type:technical-enabler" = @("Technical Enabler", "Acceptance Criteria")
        "type:security"          = @("Security", "Acceptance Criteria")
        "type:qa"                = @("QA")
        "type:doc"               = @("Documentation", "Acceptance Criteria")
        "type:chore"             = @("Maintenance", "Acceptance Criteria")
    }

    $RequiredAnySectionByType = @{
        "type:bug" = @("Bug Report", "Problem Summary")
        "type:qa"  = @("Validation", "Acceptance Criteria", "Scenarios", "Pass Conditions")
    }

    foreach ($RequiredSection in $RequiredSectionsByType[$CardType]) {
        if (-not (Test-BodyContains -Body $Body -Needle $RequiredSection)) {
            return $false
        }
    }

    if ($RequiredAnySectionByType.ContainsKey($CardType)) {
        $FoundAny = $false
        foreach ($RequiredSection in $RequiredAnySectionByType[$CardType]) {
            if (Test-BodyContains -Body $Body -Needle $RequiredSection) {
                $FoundAny = $true
                break
            }
        }

        if (-not $FoundAny) {
            return $false
        }
    }

    return $true
}

function Get-ProjectFields {
    return Invoke-GhJson -Arguments @(
        "project", "field-list", [string]$script:ProjectNumber,
        "--owner", $script:ProjectOwner,
        "--format", "json"
    )
}

function Find-ProjectField {
    param(
        [Parameter(Mandatory)][object[]]$Fields,
        [Parameter(Mandatory)][string[]]$Names
    )

    foreach ($Name in $Names) {
        $Match = @($Fields | Where-Object { $_.name -ieq $Name })
        if ($Match.Count -gt 1) {
            throw "Multiple project fields matched '$Name'. Resolve the duplicate fields before running the script."
        }

        if ($Match.Count -eq 1) {
            return $Match[0]
        }
    }

    return $null
}

function Test-OptionNameOrder {
    param(
        [AllowNull()][object[]]$CurrentOptions,
        [Parameter(Mandatory)][object[]]$DesiredOptions
    )

    $CurrentNames = @($CurrentOptions | ForEach-Object { [string]$_.name })
    $DesiredNames = @($DesiredOptions | ForEach-Object { [string]$_.name })

    if ($CurrentNames.Count -ne $DesiredNames.Count) {
        return $false
    }

    for ($Index = 0; $Index -lt $DesiredNames.Count; $Index++) {
        if ($CurrentNames[$Index] -cne $DesiredNames[$Index]) {
            return $false
        }
    }

    return $true
}

function Merge-OptionDefinitions {
    param(
        [Parameter(Mandatory)][object[]]$DesiredOptions,
        [AllowNull()][object[]]$CurrentOptions
    )

    $Merged = New-Object System.Collections.Generic.List[object]
    $DesiredNames = @{}

    foreach ($Option in $DesiredOptions) {
        $Merged.Add([pscustomobject]@{
            name        = [string]$Option.name
            color       = [string]$Option.color
            description = [string]$Option.description
        }) | Out-Null
        $DesiredNames[[string]$Option.name] = $true
    }

    foreach ($Option in @($CurrentOptions)) {
        $Name = [string]$Option.name
        if ([string]::IsNullOrWhiteSpace($Name) -or $DesiredNames.ContainsKey($Name)) {
            continue
        }

        $Color = if ($Option.PSObject.Properties.Name -contains "color" -and $Option.color) {
            [string]$Option.color
        }
        else {
            "GRAY"
        }

        $Description = if ($Option.PSObject.Properties.Name -contains "description" -and $Option.description) {
            [string]$Option.description
        }
        else {
            "Legacy option retained temporarily while cards are migrated"
        }

        $Merged.Add([pscustomobject]@{
            name        = $Name
            color       = $Color
            description = $Description
        }) | Out-Null
    }

    # Windows PowerShell 5.1 can throw "Argument types do not match" when
    # a generic List[object] is wrapped with @(...). Convert it explicitly.
    return $Merged.ToArray()
}

function Set-SingleSelectFieldOptions {
    param(
        [Parameter(Mandatory)][string]$FieldId,
        [Parameter(Mandatory)][object[]]$Options,
        [Parameter(Mandatory)][string]$FieldName
    )

    if ($DryRun) {
        Write-Host "DRY-RUN: would set '$FieldName' options to: $((@($Options.name)) -join ', ')" -ForegroundColor Yellow
        return
    }

    $Mutation = @'
mutation($input: UpdateProjectV2FieldInput!) {
  updateProjectV2Field(input: $input) {
    projectV2Field {
      ... on ProjectV2SingleSelectField {
        id
        name
        options { id name }
      }
    }
  }
}
'@

    $Request = @{
        query = $Mutation
        variables = @{
            input = @{
                fieldId = $FieldId
                singleSelectOptions = @($Options | ForEach-Object {
                    @{
                        name = [string]$_.name
                        color = [string]$_.color
                        description = [string]$_.description
                    }
                })
            }
        }
    } | ConvertTo-Json -Depth 12 -Compress

    $Request | & gh api graphql --input - | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to update single-select options for project field '$FieldName'."
    }
}

function Ensure-SingleSelectFieldForMigration {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][object[]]$DesiredOptions,
        [switch]$BuiltInStatus
    )

    $FieldsResult = Get-ProjectFields
    $Fields = @($FieldsResult.fields)
    $FieldNames = if ($BuiltInStatus) { @("Status", "Workflow Status") } else { @($Name) }
    $Field = Find-ProjectField -Fields $Fields -Names $FieldNames
    $Created = $false

    if (-not $Field) {
        if ($BuiltInStatus) {
            throw "The project does not contain its built-in Status field."
        }

        if ($DryRun) {
            Write-Host "DRY-RUN: would create project field '$Name'." -ForegroundColor Yellow
            return [pscustomobject]@{ id = "dry-run-$Name"; name = $Name; options = @() }
        }

        $Arguments = @(
            "project", "field-create", [string]$script:ProjectNumber,
            "--owner", $script:ProjectOwner,
            "--name", $Name,
            "--data-type", "SINGLE_SELECT"
        )

        foreach ($Option in $DesiredOptions) {
            $Arguments += @("--single-select-options", [string]$Option.name)
        }

        Invoke-GhCommand -Arguments $Arguments

        $FieldsResult = Get-ProjectFields
        $Fields = @($FieldsResult.fields)
        $Field = Find-ProjectField -Fields $Fields -Names @($Name)
        if (-not $Field) {
            throw "Project field '$Name' was created but could not be retrieved."
        }
        $Created = $true
    }

    $MigrationOptions = Merge-OptionDefinitions -DesiredOptions $DesiredOptions -CurrentOptions @($Field.options)
    if ($Created -or -not (Test-OptionNameOrder -CurrentOptions @($Field.options) -DesiredOptions $MigrationOptions)) {
        Set-SingleSelectFieldOptions -FieldId ([string]$Field.id) -Options $MigrationOptions -FieldName ([string]$Field.name)
    }
    else {
        Write-Host "Field '$($Field.name)' already contains the required migration options." -ForegroundColor DarkGreen
    }

    return $Field
}

function Finalize-SingleSelectField {
    param(
        [Parameter(Mandatory)][string[]]$Names,
        [Parameter(Mandatory)][object[]]$DesiredOptions
    )

    $FieldsResult = Get-ProjectFields
    $Field = Find-ProjectField -Fields @($FieldsResult.fields) -Names $Names
    if (-not $Field) {
        throw "Project field not found during finalization: $($Names -join ', ')."
    }

    if (Test-OptionNameOrder -CurrentOptions @($Field.options) -DesiredOptions $DesiredOptions) {
        Write-Host "Field '$($Field.name)' already has the final approved options." -ForegroundColor DarkGreen
        return
    }

    Set-SingleSelectFieldOptions -FieldId ([string]$Field.id) -Options $DesiredOptions -FieldName ([string]$Field.name)
}

function Get-OptionIdMap {
    param([Parameter(Mandatory)][object]$Field)

    $Map = @{}
    foreach ($Option in @($Field.options)) {
        $Map[[string]$Option.name] = [string]$Option.id
    }
    return $Map
}

function Ensure-RepositoryLabels {
    $Existing = Invoke-GhJson -Arguments @(
        "label", "list",
        "--repo", $script:Repo,
        "--limit", "1000",
        "--json", "name"
    )

    $ExistingNames = @($Existing | ForEach-Object { [string]$_.name })

    foreach ($Label in $LabelsToEnsure) {
        if ($ExistingNames -contains $Label.Name) {
            continue
        }

        if ($DryRun) {
            Write-Host "DRY-RUN: would create label '$($Label.Name)'." -ForegroundColor Yellow
            continue
        }

        Invoke-GhCommand -Arguments @(
            "label", "create", $Label.Name,
            "--repo", $script:Repo,
            "--color", $Label.Color,
            "--description", $Label.Description
        )
    }
}

function Get-OpenIssues {
    $Result = Invoke-GhJson -Arguments @(
        "issue", "list",
        "--repo", $script:Repo,
        "--state", "open",
        "--limit", "1000",
        "--json", "number,title,url,body,labels"
    )

    return @($Result)
}

function Get-OpenIssueByExactTitle {
    param(
        [Parameter(Mandatory)][object[]]$OpenIssues,
        [Parameter(Mandatory)][string]$Title
    )

    $Matches = @($OpenIssues | Where-Object { $_.title -ceq $Title })
    if ($Matches.Count -gt 1) {
        throw "Multiple open issues have the exact title '$Title'. Resolve the duplicate before running the script."
    }

    if ($Matches.Count -eq 0) {
        return $null
    }

    return $Matches[0]
}

function Get-ProjectItems {
    $Result = Invoke-GhJson -Arguments @(
        "project", "item-list", [string]$script:ProjectNumber,
        "--owner", $script:ProjectOwner,
        "--limit", [string]$ProjectLimit,
        "--format", "json"
    )

    if ($null -eq $Result) {
        return @()
    }

    if ($Result.PSObject.Properties.Name -contains "items") {
        return @($Result.items)
    }

    return @($Result)
}

function Get-ProjectItemForIssue {
    param(
        [Parameter(Mandatory)][object[]]$ProjectItems,
        [Parameter(Mandatory)][object]$Issue
    )

    $Matches = New-Object System.Collections.Generic.List[object]

    foreach ($Item in $ProjectItems) {
        $Content = if ($Item.PSObject.Properties.Name -contains "content") { $Item.content } else { $null }
        if ($null -eq $Content) {
            continue
        }

        $UrlMatches = ($Content.PSObject.Properties.Name -contains "url") -and ([string]$Content.url -eq [string]$Issue.url)

        $RepositoryMatches = $true
        if ($Content.PSObject.Properties.Name -contains "repository") {
            $RepositoryValue = $Content.repository
            if ($RepositoryValue -is [string]) {
                $RepositoryMatches = ([string]$RepositoryValue -ieq $script:Repo)
            }
            elseif ($RepositoryValue -and ($RepositoryValue.PSObject.Properties.Name -contains "nameWithOwner")) {
                $RepositoryMatches = ([string]$RepositoryValue.nameWithOwner -ieq $script:Repo)
            }
        }

        $NumberMatches = $RepositoryMatches -and
            ($Content.PSObject.Properties.Name -contains "number") -and
            ([int]$Content.number -eq [int]$Issue.number)

        if ($UrlMatches -or $NumberMatches) {
            $Matches.Add($Item) | Out-Null
        }
    }

    if ($Matches.Count -gt 1) {
        throw "Issue #$($Issue.number) appears more than once in the project. Remove the duplicate project item before continuing."
    }

    if ($Matches.Count -eq 0) {
        return $null
    }

    return $Matches[0]
}

function Ensure-ProjectItem {
    param(
        [Parameter(Mandatory)][object]$Issue,
        [Parameter(Mandatory)][object[]]$ProjectItems
    )

    $Existing = Get-ProjectItemForIssue -ProjectItems $ProjectItems -Issue $Issue
    if ($Existing) {
        return $Existing
    }

    if ($DryRun) {
        Write-Host "DRY-RUN: would add issue #$($Issue.number) to project '$ProjectTitle'." -ForegroundColor Yellow
        return [pscustomobject]@{ id = "dry-run-item-$($Issue.number)"; content = $Issue }
    }

    $Added = Invoke-GhJson -Arguments @(
        "project", "item-add", [string]$script:ProjectNumber,
        "--owner", $script:ProjectOwner,
        "--url", [string]$Issue.url,
        "--format", "json"
    )

    if (-not $Added -or -not $Added.id) {
        throw "Issue #$($Issue.number) was added to the project but no project item ID was returned."
    }

    return $Added
}

function Get-LabelNames {
    param([Parameter(Mandatory)][object]$Issue)

    return @($Issue.labels | ForEach-Object { [string]$_.name })
}

function Sync-IssueLabels {
    param(
        [Parameter(Mandatory)][object]$Issue,
        [Parameter(Mandatory)][string[]]$DesiredLabels,
        [Parameter(Mandatory)][string]$WorkflowStatus
    )

    $CurrentLabels = Get-LabelNames -Issue $Issue
    $EffectiveDesired = New-Object System.Collections.Generic.List[string]
    foreach ($Label in $DesiredLabels) {
        if (-not $EffectiveDesired.Contains($Label)) {
            $EffectiveDesired.Add($Label) | Out-Null
        }
    }

    if ($WorkflowStatus -eq "Needs Refinement") {
        if (-not $EffectiveDesired.Contains("needs-refinement")) {
            $EffectiveDesired.Add("needs-refinement") | Out-Null
        }
    }

    $Families = @("type:", "priority:", "scope:")
    $ToRemove = New-Object System.Collections.Generic.List[string]

    foreach ($Current in $CurrentLabels) {
        $IsConflictingFamilyLabel = $false
        foreach ($Prefix in $Families) {
            if ($Current.StartsWith($Prefix, [System.StringComparison]::OrdinalIgnoreCase) -and -not $EffectiveDesired.Contains($Current)) {
                $IsConflictingFamilyLabel = $true
                break
            }
        }

        if ($IsConflictingFamilyLabel) {
            $ToRemove.Add($Current) | Out-Null
        }
    }

    if ($WorkflowStatus -ne "Needs Refinement" -and $CurrentLabels -contains "needs-refinement") {
        $ToRemove.Add("needs-refinement") | Out-Null
    }

    $ToAdd = @($EffectiveDesired | Where-Object { $CurrentLabels -notcontains $_ })

    if ($DryRun) {
        if ($ToRemove.Count -gt 0) {
            Write-Host "DRY-RUN: issue #$($Issue.number) remove labels: $($ToRemove -join ', ')" -ForegroundColor Yellow
        }
        if ($ToAdd.Count -gt 0) {
            Write-Host "DRY-RUN: issue #$($Issue.number) add labels: $($ToAdd -join ', ')" -ForegroundColor Yellow
        }
        return
    }

    if ($ToRemove.Count -gt 0) {
        Invoke-GhCommand -Arguments @(
            "issue", "edit", [string]$Issue.number,
            "--repo", $script:Repo,
            "--remove-label", ($ToRemove -join ",")
        )
    }

    if ($ToAdd.Count -gt 0) {
        Invoke-GhCommand -Arguments @(
            "issue", "edit", [string]$Issue.number,
            "--repo", $script:Repo,
            "--add-label", ($ToAdd -join ",")
        )
    }
}

function Set-ProjectItemSingleSelect {
    param(
        [Parameter(Mandatory)][string]$ItemId,
        [Parameter(Mandatory)][string]$FieldId,
        [Parameter(Mandatory)][hashtable]$OptionMap,
        [Parameter(Mandatory)][string]$Value,
        [Parameter(Mandatory)][string]$FieldName
    )

    if (-not $OptionMap.ContainsKey($Value)) {
        throw "Project field '$FieldName' does not contain option '$Value'."
    }

    if ($DryRun) {
        Write-Host "DRY-RUN: would set project item '$ItemId' $FieldName='$Value'." -ForegroundColor Yellow
        return
    }

    Invoke-GhCommand -Arguments @(
        "project", "item-edit",
        "--id", $ItemId,
        "--project-id", $script:ProjectId,
        "--field-id", $FieldId,
        "--single-select-option-id", $OptionMap[$Value]
    )
}

function Invoke-SelfTest {
    $Failures = New-Object System.Collections.Generic.List[string]

    function Assert-Test {
        param([string]$Name, [bool]$Condition)
        if ($Condition) {
            Write-Host "PASS: $Name" -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: $Name" -ForegroundColor Red
            $Failures.Add($Name) | Out-Null
        }
    }

    $ValidBugBody = @"
## Problem Summary

The worker image omits a required runtime module.

## Acceptance Criteria

- The image builds.
- The import resolves.
"@

    $ValidStoryBody = @"
## User Story

As a creator, I need to upload a ZIP so that a batch can be processed.

## Acceptance Criteria

- A valid ZIP creates a batch.
"@

    Assert-Test "empty body rejected" (-not (Test-StructuredIssueBody -Body "" -Labels @("type:bug")))
    Assert-Test "old boilerplate rejected" (-not (Test-StructuredIssueBody -Body "## Agile Rule" -Labels @("type:bug")))
    Assert-Test "valid bug accepted" (Test-StructuredIssueBody -Body $ValidBugBody -Labels @("type:bug"))
    Assert-Test "valid story accepted" (Test-StructuredIssueBody -Body $ValidStoryBody -Labels @("type:user-story"))
    Assert-Test "multiple type labels rejected" (-not (Test-StructuredIssueBody -Body $ValidBugBody -Labels @("type:bug", "type:qa")))
    Assert-Test "approved status count" ($DesiredStatusOptions.Count -eq 8)
    Assert-Test "QA review status present" (@($DesiredStatusOptions.name) -contains "QA / In Review")
    Assert-Test "blocked status present" (@($DesiredStatusOptions.name) -contains "Blocked")

    if ($Failures.Count -gt 0) {
        throw "Self-test failed: $($Failures -join ', ')"
    }

    Write-Host "Self-test complete: all checks passed." -ForegroundColor Green
}

if ($SelfTest) {
    Invoke-SelfTest
    return
}

Write-Section "Authentication and repository"
Invoke-GhCommand -Arguments @("auth", "status")

$script:Repo = (& gh repo view --json nameWithOwner -q ".nameWithOwner").Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($script:Repo)) {
    throw "Could not resolve the current GitHub repository. Run this script from inside the ClipSense repository."
}

$script:ProjectOwner = (& gh repo view --json owner -q ".owner.login").Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($script:ProjectOwner)) {
    throw "Could not resolve the repository owner."
}

Write-Host "Repository: $script:Repo" -ForegroundColor Green
Write-Host "Project owner: $script:ProjectOwner" -ForegroundColor Green

Write-Section "Resolve project"
$ProjectsResult = Invoke-GhJson -Arguments @(
    "project", "list",
    "--owner", $script:ProjectOwner,
    "--limit", "100",
    "--format", "json"
)

$ProjectMatches = @($ProjectsResult.projects | Where-Object { $_.title -ieq $ProjectTitle })
if ($ProjectMatches.Count -eq 0) {
    throw "Project '$ProjectTitle' was not found under '$script:ProjectOwner'."
}
if ($ProjectMatches.Count -gt 1) {
    throw "Multiple projects match '$ProjectTitle'. Rename or close the duplicate before running the script."
}

$Project = $ProjectMatches[0]
$script:ProjectNumber = [int]$Project.number
$ProjectView = Invoke-GhJson -Arguments @(
    "project", "view", [string]$script:ProjectNumber,
    "--owner", $script:ProjectOwner,
    "--format", "json"
)
$script:ProjectId = [string]$ProjectView.id

Write-Host "Project: $($Project.title) #$script:ProjectNumber" -ForegroundColor Green
Write-Host "Project ID: $script:ProjectId" -ForegroundColor DarkGreen

Write-Section "Repository labels"
Ensure-RepositoryLabels
Write-Host "Required issue labels verified." -ForegroundColor Green

Write-Section "Prepare project fields"
$StatusField = Ensure-SingleSelectFieldForMigration -Name "Status" -DesiredOptions $DesiredStatusOptions -BuiltInStatus
$PriorityField = Ensure-SingleSelectFieldForMigration -Name "Priority" -DesiredOptions $DesiredPriorityOptions
$TypeField = Ensure-SingleSelectFieldForMigration -Name "Type" -DesiredOptions $DesiredTypeOptions
$ScopeField = Ensure-SingleSelectFieldForMigration -Name "Scope" -DesiredOptions $DesiredScopeOptions

if (-not $DryRun) {
    $FieldsResult = Get-ProjectFields
    $AllFields = @($FieldsResult.fields)
    $StatusField = Find-ProjectField -Fields $AllFields -Names @("Status", "Workflow Status")
    $PriorityField = Find-ProjectField -Fields $AllFields -Names @("Priority")
    $TypeField = Find-ProjectField -Fields $AllFields -Names @("Type")
    $ScopeField = Find-ProjectField -Fields $AllFields -Names @("Scope")

    foreach ($RequiredField in @($StatusField, $PriorityField, $TypeField, $ScopeField)) {
        if (-not $RequiredField) {
            throw "A required project field could not be resolved after configuration."
        }
    }

    $StatusOptionIds = Get-OptionIdMap -Field $StatusField
    $PriorityOptionIds = Get-OptionIdMap -Field $PriorityField
    $TypeOptionIds = Get-OptionIdMap -Field $TypeField
    $ScopeOptionIds = Get-OptionIdMap -Field $ScopeField
}
else {
    $StatusOptionIds = @{}
    $PriorityOptionIds = @{}
    $TypeOptionIds = @{}
    $ScopeOptionIds = @{}
}

$Cards = @(
    @{ Title = "US-000: Transform Raw Creator Footage Into Story-Ready Output"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:user-story","priority:p0","scope:mvp") }
    @{ Title = "US-001: ZIP Batch Upload"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:user-story","priority:p0","scope:mvp") }
    @{ Title = "Safe ZIP extraction for worker processing"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:technical-enabler","priority:p0","scope:mvp") }
    @{ Title = "US-003: AI Video Processing"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:user-story","priority:p1","scope:mvp") }
    @{ Title = "US-004: Batch Result Review"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:user-story","priority:p0","scope:mvp") }
    @{ Title = "US-005: Clear Processing Failure Reasons"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:user-story","priority:p1","scope:mvp") }
    @{ Title = "Poison-pill worker job handling"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "Disk cleanup for large media processing"; WorkflowStatus = "Product Backlog"; Priority = "P2"; Labels = @("type:chore","priority:p2","scope:stabilization") }
    @{ Title = "Full local runtime verification"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:qa","priority:p0","scope:mvp") }
    @{ Title = "US-009: Upload And API Error Clarity"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:user-story","priority:p1","scope:mvp") }
    @{ Title = "US-010: Authenticated Export Download"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:user-story","priority:p1","scope:mvp") }

    @{ Title = "CS-101: Provide a reproducible local environment for the full ClipSense stack"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:technical-enabler","priority:p0","scope:mvp") }
    @{ Title = "CS-114: Local Docker Runtime Repair and Verification"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:chore","priority:p0","scope:mvp") }
    @{ Title = "CS-110: Worker Docker Packaging Fix"; WorkflowStatus = "Ready"; Priority = "P0"; Labels = @("type:bug","priority:p0","scope:mvp") }
    @{ Title = "CS-102: Define and verify the MVP Postgres persistence baseline"; WorkflowStatus = "Needs Refinement"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp","needs-refinement") }
    @{ Title = "CS-107: Local JWT Auth and Protected Batch Routes"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-108: Local Browser CORS Policy"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:security","priority:p1","scope:mvp") }
    @{ Title = "CS-109: JWT Secret Configuration and Production Guardrails"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:security","priority:p1","scope:stabilization") }
    @{ Title = "CS-103: Go API Multipart ZIP Intake"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:technical-enabler","priority:p0","scope:mvp") }
    @{ Title = "CS-111: Upload Size and ZIP Validation Hardening"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:security","priority:p0","scope:mvp") }
    @{ Title = "CS-112: API Error Response Semantics"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:bug","priority:p1","scope:mvp") }
    @{ Title = "CS-201: Queue uploaded batches for asynchronous worker processing"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:technical-enabler","priority:p0","scope:mvp") }
    @{ Title = "CS-104: Python Worker Safe Extraction and Audio Prep Pipeline"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:technical-enabler","priority:p0","scope:mvp") }
    @{ Title = "CS-203: Python Worker Transcription Integration"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-205: Basic MVP Clip Classification and Taxonomy Foundation"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-204: Store clip embeddings in Qdrant for semantic workflows"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-207: Persist and Display Batch Failure Reasons"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-301: MVP Creator Dashboard and Batch Review Views"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:technical-enabler","priority:p0","scope:mvp") }
    @{ Title = "CS-305: Fix Dashboard Route Structure"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:bug","priority:p0","scope:mvp") }
    @{ Title = "CS-306: Stabilize Frontend Auth Hydration"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:bug","priority:p0","scope:mvp") }
    @{ Title = "CS-304: Show reliable frontend errors for API and processing failures"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-405: MVP Batch Export and Future Narrative Package Export"; WorkflowStatus = "Product Backlog"; Priority = "P2"; Labels = @("type:technical-enabler","priority:p2","scope:mvp") }
    @{ Title = "CS-113: Frontend Dependency Vulnerability Cleanup"; WorkflowStatus = "Needs Refinement"; Priority = "P1"; Labels = @("type:security","priority:p1","scope:stabilization","needs-refinement") }
    @{ Title = "CS-115: CI Build/Test Validation"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:qa","priority:p1","scope:stabilization") }
    @{ Title = "CS-208: End-to-End ZIP MVP Verification Gate"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:qa","priority:p0","scope:mvp") }
    @{ Title = "CS-117: Stabilization Documentation Refresh"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:doc","priority:p1","scope:stabilization") }
    @{ Title = "CS-105: Postgres Connection Pooling and Health-Check Handshake"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:stabilization") }
    @{ Title = "CS-106: Disk-Space Janitor and TTL Cleanup Daemon"; WorkflowStatus = "Needs Refinement"; Priority = "P2"; Labels = @("type:technical-enabler","priority:p2","scope:stabilization","needs-refinement") }
    @{ Title = "CS-116: Replace Legacy SQLite Backup and Restore Scripts"; WorkflowStatus = "Product Backlog"; Priority = "P2"; Labels = @("type:chore","priority:p2","scope:stabilization") }
    @{ Title = "CS-119: Python Dependency Audit Path"; WorkflowStatus = "Needs Refinement"; Priority = "P2"; Labels = @("type:security","priority:p2","scope:stabilization","needs-refinement") }
    @{ Title = "CS-202: Go API Progress Broadcaster"; WorkflowStatus = "Needs Refinement"; Priority = "P2"; Labels = @("type:technical-enabler","priority:p2","scope:roadmap","needs-refinement") }
    @{ Title = "CS-206: Poison Pill Queue Catch-All and Dead Letter Queue"; WorkflowStatus = "Needs Refinement"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp","needs-refinement") }

    @{ Title = "US-011: Direct Video Upload"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:roadmap") }
    @{ Title = "US-012: Public Or User-Provided Link Intake"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:roadmap") }
    @{ Title = "US-013: Video Segment Preview"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:post-mvp") }
    @{ Title = "US-014: Timeline Story Workspace"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:post-mvp") }
    @{ Title = "US-015: Advanced Narrative Intelligence"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:post-mvp") }
    @{ Title = "US-016: Scaled Cloud Uploads"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:post-mvp") }
    @{ Title = "US-017: Faster Video Processing Engine"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:post-mvp") }
    @{ Title = "US-018: Editor-Ready Export Package"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:user-story","priority:p3","scope:post-mvp") }
    @{ Title = "CS-302: Native Video Segment Player"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-303: Timeline Workspace State Store"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-401: Remote Video Link Ingestion Worker"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:roadmap") }
    @{ Title = "CS-402: Cloud Multipart Upload Pipeline"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-403: Graph-Based Narrative Connector"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-404: Rust Video Core for Scene Boundary Detection"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "Advanced scope of CS-405: Editor-Ready Export Package"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
)


Write-Section "Load current issues and project items"
$OpenIssues = Get-OpenIssues
$ProjectItems = Get-ProjectItems
Write-Host "Open issues available: $($OpenIssues.Count)" -ForegroundColor Green
Write-Host "Existing project items: $($ProjectItems.Count)" -ForegroundColor Green

$PlacedCount = 0
$SkippedCount = 0
$EffectiveStatusCounts = @{}

Write-Section "Synchronize cards"
foreach ($Card in $Cards) {
    $Issue = Get-OpenIssueByExactTitle -OpenIssues $OpenIssues -Title $Card.Title
    if (-not $Issue) {
        Write-Warning "Skipped missing open issue: $($Card.Title)"
        $SkippedCount++
        continue
    }

    $EffectiveStatus = [string]$Card.WorkflowStatus
    if (-not (Test-StructuredIssueBody -Body ([string]$Issue.body) -Labels @($Card.Labels))) {
        $EffectiveStatus = "Needs Refinement"
        Write-Warning "Issue #$($Issue.number) failed body/type validation and will be placed in Needs Refinement instead of '$($Card.WorkflowStatus)'."
    }

    $TypeLabel = Get-CardTypeLabel -Labels @($Card.Labels)
    $ScopeLabel = Get-CardScopeLabel -Labels @($Card.Labels)
    if (-not $TypeLabel -or -not $ScopeLabel) {
        Write-Warning "Skipped issue #$($Issue.number): exactly one approved type label and one approved scope label are required."
        $SkippedCount++
        continue
    }

    Sync-IssueLabels -Issue $Issue -DesiredLabels @($Card.Labels) -WorkflowStatus $EffectiveStatus
    $ProjectItem = Ensure-ProjectItem -Issue $Issue -ProjectItems $ProjectItems

    if (-not $DryRun) {
        Set-ProjectItemSingleSelect -ItemId ([string]$ProjectItem.id) -FieldId ([string]$StatusField.id) -OptionMap $StatusOptionIds -Value $EffectiveStatus -FieldName "Status"
        Set-ProjectItemSingleSelect -ItemId ([string]$ProjectItem.id) -FieldId ([string]$PriorityField.id) -OptionMap $PriorityOptionIds -Value ([string]$Card.Priority) -FieldName "Priority"
        Set-ProjectItemSingleSelect -ItemId ([string]$ProjectItem.id) -FieldId ([string]$TypeField.id) -OptionMap $TypeOptionIds -Value $TypeLabelToFieldValue[$TypeLabel] -FieldName "Type"
        Set-ProjectItemSingleSelect -ItemId ([string]$ProjectItem.id) -FieldId ([string]$ScopeField.id) -OptionMap $ScopeOptionIds -Value $ScopeLabelToFieldValue[$ScopeLabel] -FieldName "Scope"
    }
    else {
        Write-Host "DRY-RUN: #$($Issue.number) '$($Issue.title)' -> $EffectiveStatus / $($Card.Priority) / $($TypeLabelToFieldValue[$TypeLabel]) / $($ScopeLabelToFieldValue[$ScopeLabel])" -ForegroundColor Yellow
    }

    if (-not $EffectiveStatusCounts.ContainsKey($EffectiveStatus)) {
        $EffectiveStatusCounts[$EffectiveStatus] = 0
    }
    $EffectiveStatusCounts[$EffectiveStatus]++
    $PlacedCount++

    Write-Host "Placed #$($Issue.number): $($Issue.title) -> $EffectiveStatus" -ForegroundColor Green
}

Write-Section "Finalize approved field options"
if ($DryRun) {
    Write-Host "DRY-RUN: would remove legacy project-field options after card migration." -ForegroundColor Yellow
}
else {
    Finalize-SingleSelectField -Names @("Status", "Workflow Status") -DesiredOptions $DesiredStatusOptions
    Finalize-SingleSelectField -Names @("Priority") -DesiredOptions $DesiredPriorityOptions
    Finalize-SingleSelectField -Names @("Type") -DesiredOptions $DesiredTypeOptions
    Finalize-SingleSelectField -Names @("Scope") -DesiredOptions $DesiredScopeOptions
}

Write-Section "Summary"
Write-Host "Project: $ProjectTitle #$script:ProjectNumber" -ForegroundColor Cyan
Write-Host "Cards configured: $PlacedCount" -ForegroundColor Green
Write-Host "Cards skipped: $SkippedCount" -ForegroundColor $(if ($SkippedCount -eq 0) { "Green" } else { "Yellow" })

foreach ($StatusName in @($DesiredStatusOptions.name)) {
    $Count = if ($EffectiveStatusCounts.ContainsKey($StatusName)) { $EffectiveStatusCounts[$StatusName] } else { 0 }
    Write-Host ("{0,-20} {1,3}" -f $StatusName, $Count)
}

if ($DryRun) {
    Write-Host "Dry run complete. No GitHub data was changed." -ForegroundColor Yellow
}
else {
    Write-Host "Kanban synchronization complete." -ForegroundColor Green
    Write-Host "Expected populated columns: Product Backlog=35, Needs Refinement=6, Ready=1, Icebox=15." -ForegroundColor DarkGreen
    Write-Host "In Progress, QA / In Review, Blocked, and Done remain empty until work moves through delivery." -ForegroundColor DarkGreen
}
