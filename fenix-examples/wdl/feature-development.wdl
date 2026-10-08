# Fénix Workflow Definition Language (WDL) Example
# Simple Feature Development Workflow

workflow FeatureDevelopment

on FeatureRequested
    pipeline FeaturePipeline

pipeline FeaturePipeline

stage Analysis
    processor DiscoveryProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 10m
            resources
                max-workers 2
                memory 2GB
                priority high

stage Design
    depends Analysis
    processor ArchitectureProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 15m
            resources
                max-workers 2
                memory 2GB
                priority high

stage Implementation
    depends Design
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

stage Testing
    depends Implementation
    processor TestProcessor
        policy
            retry
                attempts 2
                backoff exponential
                timeout 20m
        resources
            max-workers 3
            memory 2GB
            priority high

stage Review
    depends Testing
    processor CodeReviewProcessor
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

stage Deployment
    depends Review
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
