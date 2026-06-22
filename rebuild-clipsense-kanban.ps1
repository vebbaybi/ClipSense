# ============================================================
# ClipSense GitHub Project Placement Script
# ============================================================
# Run inside the ClipSense repo.
#
# This script places existing refined issues into a GitHub Project.
# It intentionally does not generate generic issue bodies.
#
# Safe behavior:
# - Does not delete, close, or reopen issues.
# - Reuses open issues by exact title.
# - Never overwrites a meaningful existing issue body.
# - Never writes workflow status or priority into issue bodies.
# - Never inserts generic Agile declarations into issue bodies.
# - Requires issue-specific body content before creating a missing issue.
# - Validates body structure against the card's work-item type.
# - Reports skipped issues and why they were skipped.
# - Keeps project placement separate from requirement content.
# ============================================================

$ErrorActionPreference = "Stop"

$ProjectTitle = "ClipSense"
$WorkflowFieldNames = @("Workflow Status", "Status")
$PreferredBacklogStatus = "Product Backlog"
$PreferredNeedsRefinementStatus = "Needs Refinement"

gh auth status | Out-Null

$Repo = gh repo view --json nameWithOwner -q ".nameWithOwner"
$ProjectOwner = gh repo view --json owner -q ".owner.login"

Write-Host "Repo: $Repo" -ForegroundColor Cyan
Write-Host "Project owner: $ProjectOwner" -ForegroundColor Cyan
Write-Host "Project title: $ProjectTitle" -ForegroundColor Cyan

$Projects = gh project list --owner $ProjectOwner --format json | ConvertFrom-Json
$Project = @($Projects.projects) |
    Where-Object { $_.title -eq $ProjectTitle } |
    Select-Object -First 1

if (-not $Project) {
    throw "Project '$ProjectTitle' not found under '$ProjectOwner'. Create the project and fields before running this placement script."
}

$ProjectNumber = $Project.number
$ProjectView = gh project view $ProjectNumber --owner $ProjectOwner --format json | ConvertFrom-Json
$ProjectId = $ProjectView.id

Write-Host "Found project: $ProjectTitle #$ProjectNumber" -ForegroundColor Green

$Fields = gh project field-list $ProjectNumber --owner $ProjectOwner --format json | ConvertFrom-Json

$WorkflowField = $null
foreach ($WorkflowFieldName in $WorkflowFieldNames) {
    $WorkflowField = @($Fields.fields) |
        Where-Object { $_.name -eq $WorkflowFieldName } |
        Select-Object -First 1

    if ($WorkflowField) {
        break
    }
}

if (-not $WorkflowField) {
    throw "Project workflow field not found. Expected one of: $($WorkflowFieldNames -join ', ')"
}

$WorkflowFieldId = $WorkflowField.id

function Get-WorkflowOptionId {
    param(
        [string[]]$Names,
        [bool]$Required = $true
    )

    foreach ($Name in $Names) {
        $Option = @($WorkflowField.options) |
            Where-Object { $_.name -eq $Name } |
            Select-Object -First 1

        if ($Option) {
            return $Option.id
        }
    }

    if ($Required) {
        throw "Missing workflow option. Tried: $($Names -join ', ')"
    }

    return $null
}

$WorkflowOptions = @{
    "New Issue"        = Get-WorkflowOptionId @("New Issue", "Intake") $false
    "Product Backlog"  = Get-WorkflowOptionId @("Product Backlog", "Backlog")
    "Needs Refinement" = Get-WorkflowOptionId @("Needs Refinement") $false
    "Ready"            = Get-WorkflowOptionId @("Ready")
    "Sprint Backlog"   = Get-WorkflowOptionId @("Sprint Backlog") $false
    "In Progress"      = Get-WorkflowOptionId @("In Progress", "In progress") $false
    "In Review"        = Get-WorkflowOptionId @("In Review", "In review", "Review") $false
    "Blocked"          = Get-WorkflowOptionId @("Blocked", "Bugged") $false
    "Done"             = Get-WorkflowOptionId @("Done") $false
    "Icebox"           = Get-WorkflowOptionId @("Icebox")
}

if (-not $WorkflowOptions["Needs Refinement"]) {
    $WorkflowOptions["Needs Refinement"] = $WorkflowOptions["Product Backlog"]
}

Write-Host "Workflow field verified: $($WorkflowField.name)" -ForegroundColor Green

$PriorityField = @($Fields.fields) |
    Where-Object { $_.name -eq "Priority" } |
    Select-Object -First 1

$PriorityFieldId = $null
$PriorityOptions = @{}

if ($PriorityField) {
    $PriorityFieldId = $PriorityField.id

    foreach ($Priority in @("P0", "P1", "P2", "P3")) {
        $Option = @($PriorityField.options) |
            Where-Object { $_.name -eq $Priority } |
            Select-Object -First 1

        if ($Option) {
            $PriorityOptions[$Priority] = $Option.id
        }
    }

    Write-Host "Priority field found." -ForegroundColor Green
}
else {
    Write-Host "Priority field not found. Priority labels will still be used." -ForegroundColor Yellow
}

$LabelsToEnsure = @(
    @{ Name = "type:user-story"; Color = "5319e7"; Description = "User or stakeholder outcome" },
    @{ Name = "type:technical-enabler"; Color = "0e8a16"; Description = "Engineering work that enables product delivery" },
    @{ Name = "type:feature"; Color = "1d76db"; Description = "Product capability or user-facing feature work" },
    @{ Name = "type:bug"; Color = "d73a4a"; Description = "Broken, failing, miswired, or defective behavior" },
    @{ Name = "type:qa"; Color = "0e8a16"; Description = "Testing, verification, smoke tests, and regression checks" },
    @{ Name = "type:security"; Color = "b60205"; Description = "Security, dependency risk, validation, or vulnerability work" },
    @{ Name = "type:doc"; Color = "0075ca"; Description = "Documentation and repo truth updates" },
    @{ Name = "type:chore"; Color = "cfd3d7"; Description = "Maintenance, setup, tooling, or operational work" },
    @{ Name = "needs-refinement"; Color = "d876e3"; Description = "Issue requires more refinement before it is ready to implement" },

    @{ Name = "priority:p0"; Color = "b60205"; Description = "Critical blocker or MVP-breaking work" },
    @{ Name = "priority:p1"; Color = "d93f0b"; Description = "High priority MVP/stabilization work" },
    @{ Name = "priority:p2"; Color = "fbca04"; Description = "Important but not immediately blocking" },
    @{ Name = "priority:p3"; Color = "c2e0c6"; Description = "Later, polish, roadmap, or post-MVP" },

    @{ Name = "scope:mvp"; Color = "0052cc"; Description = "Required for stabilized MVP" },
    @{ Name = "scope:stabilization"; Color = "5319e7"; Description = "MVP stabilization and hardening" },
    @{ Name = "scope:roadmap"; Color = "c2e0c6"; Description = "Planned broader product MVP work" },
    @{ Name = "scope:post-mvp"; Color = "eeeeee"; Description = "Later product expansion outside current MVP" }
)

$ExistingLabels = gh label list --repo $Repo --limit 1000 --json name | ConvertFrom-Json

foreach ($Label in $LabelsToEnsure) {
    $Found = @($ExistingLabels) |
        Where-Object { $_.name -eq $Label.Name } |
        Select-Object -First 1

    if (-not $Found) {
        gh label create $Label.Name `
            --repo $Repo `
            --color $Label.Color `
            --description $Label.Description | Out-Null
    }
}

Write-Host "Labels verified." -ForegroundColor Green

function Get-CardType {
    param([string[]]$Labels)

    $TypeLabels = @(
        "type:user-story",
        "type:bug",
        "type:technical-enabler",
        "type:security",
        "type:qa",
        "type:doc",
        "type:chore",
        "type:feature"
    )

    foreach ($TypeLabel in $TypeLabels) {
        if ($Labels -contains $TypeLabel) {
            return $TypeLabel
        }
    }

    return $null
}

function Test-StructuredIssueBody {
    param(
        [string]$Body,
        [string[]]$Labels
    )

    if ([string]::IsNullOrWhiteSpace($Body)) {
        return $false
    }

    $BoilerplateMarkers = @(
        "## Kanban Placement",
        "## Agile Rule",
        "Repository evidence can inform this card"
    )

    foreach ($Marker in $BoilerplateMarkers) {
        if ($Body.Contains($Marker)) {
            return $false
        }
    }

    $CardType = Get-CardType -Labels $Labels
    $RequiredSectionsByType = @{
        "type:user-story"        = @("## User Story", "Acceptance Criteria")
        "type:bug"               = @("Acceptance Criteria")
        "type:technical-enabler" = @("Technical Enabler", "Acceptance Criteria")
        "type:security"          = @("Security", "Acceptance Criteria")
        "type:qa"                = @("QA", "Validation")
        "type:doc"               = @("Documentation", "Acceptance Criteria")
        "type:chore"             = @("Maintenance", "Acceptance Criteria")
        "type:feature"           = @("Product", "Acceptance Criteria")
    }
    $RequiredAnySectionByType = @{
        "type:bug" = @("Bug Report", "Problem Summary")
    }

    if (-not $CardType -or -not $RequiredSectionsByType.ContainsKey($CardType)) {
        return ($Body.Length -ge 120 -and ($Body.Contains("Acceptance Criteria") -or $Body.Contains("Validation")))
    }

    foreach ($RequiredSection in $RequiredSectionsByType[$CardType]) {
        if (-not $Body.Contains($RequiredSection)) {
            return $false
        }
    }

    if ($RequiredAnySectionByType.ContainsKey($CardType)) {
        $FoundAnyRequiredSection = $false

        foreach ($RequiredSection in $RequiredAnySectionByType[$CardType]) {
            if ($Body.Contains($RequiredSection)) {
                $FoundAnyRequiredSection = $true
                break
            }
        }

        if (-not $FoundAnyRequiredSection) {
            return $false
        }
    }

    return $true
}

function Get-OpenIssueByExactTitle {
    param([string]$Title)

    $OpenIssues = gh issue list `
        --repo $Repo `
        --state open `
        --limit 1000 `
        --json number,title,url,body | ConvertFrom-Json

    return @($OpenIssues) |
        Where-Object { $_.title -eq $Title } |
        Select-Object -First 1
}

function Sync-IssueLabels {
    param(
        [int]$IssueNumber,
        [string[]]$Labels
    )

    if ($Labels.Count -gt 0) {
        gh issue edit $IssueNumber `
            --repo $Repo `
            --add-label ($Labels -join ",") | Out-Null
    }
}

function Get-OrCreate-OpenIssue {
    param(
        [string]$Title,
        [string]$Body,
        [string[]]$Labels
    )

    $Existing = Get-OpenIssueByExactTitle -Title $Title

    if ($Existing) {
        Sync-IssueLabels -IssueNumber $Existing.number -Labels $Labels
        return $Existing
    }

    if (-not (Test-StructuredIssueBody -Body $Body -Labels $Labels)) {
        Write-Warning "Skipped missing issue '$Title': no issue-specific structured body was provided for its work-item type."
        return $null
    }

    $Args = @(
        "issue", "create",
        "--repo", $Repo,
        "--title", $Title,
        "--body", $Body
    )

    foreach ($Label in $Labels) {
        $Args += "--label"
        $Args += $Label
    }

    $IssueUrl = gh @Args
    return Get-OpenIssueByExactTitle -Title $Title
}

function Add-ToProject {
    param(
        [string]$IssueUrl,
        [string]$WorkflowStatus,
        [string]$Priority
    )

    if ([string]::IsNullOrWhiteSpace($IssueUrl)) {
        return
    }

    $ProjectItem = gh project item-add $ProjectNumber `
        --owner $ProjectOwner `
        --url $IssueUrl `
        --format json | ConvertFrom-Json

    gh project item-edit `
        --id $ProjectItem.id `
        --project-id $ProjectId `
        --field-id $WorkflowFieldId `
        --single-select-option-id $WorkflowOptions[$WorkflowStatus] | Out-Null

    if ($PriorityFieldId -and $PriorityOptions.ContainsKey($Priority)) {
        gh project item-edit `
            --id $ProjectItem.id `
            --project-id $ProjectId `
            --field-id $PriorityFieldId `
            --single-select-option-id $PriorityOptions[$Priority] | Out-Null
    }
}

function Add-Card {
    param(
        [string]$Title,
        [string]$WorkflowStatus,
        [string]$Priority,
        [string[]]$Labels,
        [string]$Body
    )

    $Issue = Get-OrCreate-OpenIssue `
        -Title $Title `
        -Body $Body `
        -Labels $Labels

    if (-not $Issue) {
        return
    }

    $EffectiveStatus = $WorkflowStatus
    $EffectiveLabels = $Labels

    if (-not (Test-StructuredIssueBody -Body $Issue.body -Labels $Labels)) {
        $EffectiveStatus = $PreferredNeedsRefinementStatus
        $EffectiveLabels = @($Labels + "needs-refinement")
        Sync-IssueLabels -IssueNumber $Issue.number -Labels $EffectiveLabels
        Write-Warning "Downgraded '$Title' to $EffectiveStatus because its body is empty, generated boilerplate, or lacks the expected issue structure."
    }

    Add-ToProject `
        -IssueUrl $Issue.url `
        -WorkflowStatus $EffectiveStatus `
        -Priority $Priority

    Write-Host "Placed: $Title -> $EffectiveStatus / $Priority" -ForegroundColor Green
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
    @{ Title = "CS-301: MVP Creator Dashboard and Batch Review Views"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:feature","priority:p0","scope:mvp") }
    @{ Title = "CS-305: Fix Dashboard Route Structure"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:bug","priority:p0","scope:mvp") }
    @{ Title = "CS-306: Stabilize Frontend Auth Hydration"; WorkflowStatus = "Product Backlog"; Priority = "P0"; Labels = @("type:bug","priority:p0","scope:mvp") }
    @{ Title = "CS-304: Show reliable frontend errors for API and processing failures"; WorkflowStatus = "Product Backlog"; Priority = "P1"; Labels = @("type:technical-enabler","priority:p1","scope:mvp") }
    @{ Title = "CS-405: MVP Batch Export and Future Narrative Package Export"; WorkflowStatus = "Product Backlog"; Priority = "P2"; Labels = @("type:feature","priority:p2","scope:mvp") }
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
    @{ Title = "CS-302: Native Video Segment Player"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:feature","priority:p3","scope:post-mvp") }
    @{ Title = "CS-303: Timeline Workspace State Store"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-401: Remote Video Link Ingestion Worker"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:roadmap") }
    @{ Title = "CS-402: Cloud Multipart Upload Pipeline"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-403: Graph-Based Narrative Connector"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "CS-404: Rust Video Core for Scene Boundary Detection"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:technical-enabler","priority:p3","scope:post-mvp") }
    @{ Title = "Advanced scope of CS-405: Editor-Ready Export Package"; WorkflowStatus = "Icebox"; Priority = "P3"; Labels = @("type:feature","priority:p3","scope:post-mvp") }
)

Write-Host ""
Write-Host "Creating/reusing issues and placing them into the project..." -ForegroundColor Cyan

foreach ($Card in $Cards) {
    Add-Card `
        -Title $Card.Title `
        -WorkflowStatus $Card.WorkflowStatus `
        -Priority $Card.Priority `
        -Labels $Card.Labels `
        -Body $Card.Body
}

Write-Host ""
Write-Host "Done. Project placement complete." -ForegroundColor Green
Write-Host "No issue bodies were generated from generic boilerplate." -ForegroundColor Green
