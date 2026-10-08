# Fénix Workflow Definition Language (WDL) Example
# Standard SaaS Development Workflow

workflow StandardSaaSDevelopment

on ProductUpdated
    pipeline MainSaaSPipeline

on PullRequestOpened
    pipeline ReviewPipeline

on ReleaseRequested
    pipeline ReleasePipeline

# Main Development Pipeline
pipeline MainSaaSPipeline

stage Discovery
    processor DiscoveryProcessor
        policy
            retry
                attempts 3
                backoff exponential
                timeout 10m
            resources
                max-workers 2
                memory 2GB
                priority high

stage Requirements
    depends Discovery
    processor RequirementsProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 15m
            resources
                max-workers 2
                memory 2GB
                priority high

stage Architecture
    depends Requirements
    processor ArchitectureProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 20m
            resources
                max-workers 2
                memory 4GB
                priority high
    gate HumanApproval
        timeout 48h
        approvers
            SolutionArchitect
            ProductOwner

stage Development
    depends Architecture
    fanout
        strategy wait-all
        processors
            BackendProcessor
                policy
                    retry
                        attempts 3
                        backoff exponential
                        timeout 30m
                    resources
                        max-workers 5
                        memory 4GB
                        priority normal
            FrontendProcessor
                policy
                    retry
                        attempts 3
                        backoff exponential
                        timeout 30m
                    resources
                        max-workers 5
                        memory 4GB
                        priority normal
            DatabaseProcessor
                policy
                    retry
                        attempts 2
                        backoff exponential
                        timeout 15m
                    resources
                        max-workers 2
                        memory 2GB
                        priority normal

stage Testing
    depends Development
    fanout
        strategy wait-all
        processors
            UnitTestProcessor
                policy
                    retry
                        attempts 2
                        backoff exponential
                        timeout 10m
                    resources
                        max-workers 3
                        memory 2GB
                        priority high
            IntegrationTestProcessor
                policy
                    retry
                        attempts 2
                        backoff exponential
                        timeout 20m
                    resources
                        max-workers 3
                        memory 4GB
                        priority high
            SecurityScanProcessor
                policy
                    retry
                        attempts 1
                        backoff linear
                        timeout 15m
                    resources
                        max-workers 2
                        memory 2GB
                        priority normal

stage QualityAssurance
    depends Testing
    fanout
        strategy quorum
        minimum 2
        processors
            CodeReviewProcessor
                policy
                    retry
                        attempts 1
                        timeout 0
                    resources
                        max-workers 1
                        memory 1GB
                        priority low
            StaticAnalysisProcessor
                policy
                    retry
                        attempts 1
                        timeout 10m
                    resources
                        max-workers 2
                        memory 2GB
                        priority normal
            ManualReviewProcessor
                policy
                    retry
                        attempts 1
                        timeout 0
                    resources
                        max-workers 1
                        memory 1GB
                        priority low
    gate HumanApproval
        timeout 24h
        approvers
            TechLead
            SeniorDeveloper

stage Deployment
    depends QualityAssurance
    processor DeploymentProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 15m
            resources
                max-workers 2
                memory 2GB
                priority high

# Pull Request Review Pipeline
pipeline ReviewPipeline

stage QuickAnalysis
    processor QuickAnalysisProcessor
        policy
            retry
                attempts 1
                timeout 5m
            resources
                max-workers 2
                memory 1GB
                priority high

stage SecurityCheck
    depends QuickAnalysis
    processor SecurityReviewProcessor
        policy
            retry
                attempts 1
                timeout 10m
            resources
                max-workers 2
                memory 2GB
                priority high

stage AutomatedTests
    depends SecurityCheck
    processor QuickTestProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 10m
            resources
                max-workers 3
                memory 2GB
                priority high

# Release Pipeline
pipeline ReleasePipeline

stage PreReleaseChecks
    processor PreReleaseCheckProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 15m
            resources
                max-workers 2
                memory 2GB
                priority high

stage VersionBump
    depends PreReleaseChecks
    processor VersionBumpProcessor
        policy
            retry
                attempts 1
                timeout 5m
            resources
                max-workers 1
                memory 1GB
                priority high

stage BuildArtifacts
    depends VersionBump
    processor BuildProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 20m
            resources
                max-workers 3
                memory 4GB
                priority high

stage ReleaseStaging
    depends BuildArtifacts
    processor StagingDeploymentProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 15m
            resources
                max-workers 2
                memory 2GB
                priority high
    gate HumanApproval
        timeout 4h
        approvers
            ReleaseManager
            ProductOwner

stage ProductionRelease
    depends ReleaseStaging
    processor ProductionDeploymentProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 20m
            resources
                max-workers 2
                memory 2GB
                priority critical

stage PostRelease
    depends ProductionRelease
    processor PostReleaseProcessor
        policy
            retry
                attempts 1
                timeout 10m
            resources
                max-workers 2
                memory 2GB
                priority normal
