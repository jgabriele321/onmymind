# Project Scratchpad

## Project Phases
1. ✓ Initial Setup and Configuration (Completed)
2. ✓ Time Calculator AI Assistant (Completed)
3. Enhanced Reminder System (Planned)
4. Render Deployment & Data Migration

## Background and Motivation
- Project is a Telegram bot (MindBot) that has been refactored from WhatsApp to phone calls using Twilio
- Need to set up proper project configuration and protection
- New Requirement: Enhanced Reminder System
  - User keeps phone on vibrate and notifications muted for productivity
  - Need a way to bypass phone's vibrate mode for important reminders
  - Current phone timer solution is limited (24h max) and doesn't store reminder context
  - Need both one-off and recurring reminder capabilities
  - International usage requires platform-agnostic communication (Telegram/WhatsApp calls preferred over regular phone calls)
  - Personal tool with hardcoded communication preferences

### Time Calculator Motivation (Phase 2)
- Need quick time conversions between 24h and 12h formats
- Complex time calculations for travel planning (multiple durations)
- International time zone conversions
- Natural language processing for time-related queries
- Common use cases:
  - Converting between time formats (14:00 → 2:00 PM)
  - Backward planning for travel (flight time → when to leave)
  - International time zone differences
  - Adding/subtracting multiple time durations

## Key Challenges and Analysis
- Need proper .gitignore to exclude sensitive files and development artifacts
- Need to implement branch protection for code safety

### Time Calculator Analysis (Phase 2)
1. Technical Requirements:
   - Integration with OpenRouter API for AI processing
   - Natural language understanding for time-related queries
   - Time zone database integration
   - Support for various time formats (12h, 24h)
   - Complex time arithmetic capabilities
   - Context awareness for user's current location

2. User Experience Considerations:
   - Simple /time command interface
   - Natural language input support
   - Clear, concise responses
   - Handle ambiguous inputs gracefully
   - Support for various question formats
   - Helpful examples in error messages

3. Potential Challenges:
   - Handling ambiguous time formats
   - Dealing with daylight saving time
   - Managing multiple time zones in single query
   - Parsing complex multi-step calculations
   - Maintaining conversation context
   - Rate limiting and API costs

### Reminder System Analysis
1. Technical Requirements:
   - Integration with Telegram bot interface
   - Ability to store reminder context and timing
   - Support for both one-off and recurring reminders
   - Must use Telegram/WhatsApp calls to bypass vibrate mode
   - Need persistent storage for reminders
   - Platform-agnostic communication system that works internationally

2. User Experience Considerations:
   - Simple command format for setting reminders
   - Clear feedback when reminders are set
   - Ability to list, modify, and delete reminders
   - Context preservation for why the reminder was set
   - Fallback communication methods if primary fails

3. Potential Challenges:
   - Integration with multiple communication platforms (Telegram and WhatsApp APIs)
   - Handling international calling restrictions and regulations
   - Managing different rate limits for different platforms
   - Ensuring reliable delivery across different countries
   - Cost optimization for international communications
   - Platform-specific limitations and features
   - Handling network connectivity issues in different regions

## High-level Task Breakdown
1. Create .gitignore file
   - Success Criteria: File created with proper exclusions for:
     - Environment files (.env)
     - Build artifacts
     - IDE/Editor specific files
     - Dependency directories
     - Logs and temporary files

2. Set up branch protection
   - Success Criteria:
     - Main branch is protected
     - Status checks must pass before merging
     - Only repository owner can bypass rules
     - Protection rules are verified through GitHub API
   - Implementation Steps:
     a. Create GitHub Actions workflow for CI ✓
     b. Configure branch protection rules
     c. Verify protection is working

### Time Calculator Implementation Plan (Phase 2)
1. OpenRouter Integration
   - Success Criteria:
     - Successfully connects to OpenRouter API
     - Handles API authentication
     - Manages rate limiting
     - Implements error handling
     - Optimizes token usage

2. Time Processing System
   - Success Criteria:
     - Accurately parses time formats (12h/24h)
     - Handles time zone conversions
     - Performs time arithmetic
     - Manages daylight saving time
     - Validates time inputs

3. Natural Language Processing
   - Success Criteria:
     - Understands various time-related queries
     - Extracts time components from text
     - Identifies calculation requirements
     - Handles multi-step calculations
     - Maintains conversation context

4. Response Formatting
   - Success Criteria:
     - Provides clear, readable responses
     - Includes relevant time formats
     - Shows calculation steps when needed
     - Handles errors gracefully
     - Gives helpful suggestions

### Reminder System Implementation Plan
1. Database Schema Design
   - Success Criteria:
     - Schema supports both one-off and recurring reminders
     - Stores reminder context, timing, and user preferences
     - Handles timezone information correctly

2. Platform Integration System
   - Success Criteria:
     - Successfully integrates with both Telegram and WhatsApp APIs
     - Implements fallback mechanism between platforms
     - Handles platform-specific rate limits and restrictions
     - Manages authentication for both services

3. Command Parser Implementation
   - Success Criteria:
     - Parses natural language time inputs
     - Supports formats like "remind me in 2 hours to check email"
     - Handles recurring patterns like "every Monday at 10am"
     - Validates and confirms user input

4. Reminder Storage System
   - Success Criteria:
     - Successfully stores reminders in database
     - Handles timezone conversions
     - Provides CRUD operations for reminders
     - Implements proper error handling

5. Reminder Execution System
   - Success Criteria:
     - Accurately triggers reminders at specified times
     - Makes calls through configured platform hierarchy
     - Implements fallback to secondary platform if primary fails
     - Handles recurring reminder rescheduling
     - Implements retry mechanism for failed calls
     - Logs communication attempts and successes/failures

6. User Interface Commands
   - Success Criteria:
     - Implements /setreminder command
     - Implements /listreminders command
     - Implements /deletereminder command
     - Provides clear feedback and error messages
     - Provides platform-specific status updates

7. Testing Suite
   - Success Criteria:
     - Unit tests for all components
     - Integration tests for full reminder flow
     - Test cases for edge cases and error conditions

## Project Status Board
- [x] Create and implement .gitignore file
- [x] Create GitHub Actions workflow for CI
- [ ] Set up branch protection rules (awaiting manual configuration)
- [ ] Test branch protection

### Time Calculator Tasks (Phase 2)
- [ ] Set up OpenRouter API integration
- [ ] Implement time parsing and validation
- [ ] Add time zone conversion functionality
- [ ] Create time arithmetic system
- [ ] Develop natural language processing
- [ ] Design response formatting
- [ ] Add error handling and suggestions
- [ ] Write comprehensive tests
- [ ] Deploy and monitor

### Reminder System Tasks
- [ ] Design and implement database schema
- [ ] Set up multi-platform integration (Telegram + WhatsApp) with hardcoded preferences
- [ ] Implement command parser
- [ ] Create reminder storage system
- [ ] Develop reminder execution system with fallback logic
- [ ] Add user interface commands
- [ ] Write comprehensive tests
- [ ] Deploy and monitor system

## Deployment Analysis & Plan

### Current State
- Application is running on Render
- Data backup has been exported via /export command
- New features added: Time Calculator
- Need to preserve existing data during deployment

### Key Challenges and Analysis
1. Data Preservation Requirements:
   - SQLite database contains user items and deleted items
   - Need to maintain data continuity
   - Backup exists but no import functionality yet

2. Deployment Considerations:
   - Need to handle environment variables (BOT_TOKEN, OPENROUTER_KEY)
   - Database persistence on Render
   - Health check endpoint for Render's monitoring
   - Production-ready configuration

3. Potential Risks:
   - Data loss during deployment
   - Service interruption
   - Environment configuration mismatches

### Implementation Plan

1. Import Feature Development
   - Success Criteria:
     - Can import data from exported JSON backup
     - Handles duplicate entries gracefully
     - Preserves timestamps
     - Provides feedback on import progress
     - Command: /import with file attachment

2. Render Configuration
   - Success Criteria:
     - Environment variables properly set
     - Database persistence configured
     - Health check endpoint responding
     - Proper build and start commands
     - Logging configured

3. Deployment Process
   - Success Criteria:
     - Zero data loss
     - Minimal downtime
     - Successful health checks
     - All features functional post-deployment

### Task Breakdown

1. Import Feature Development
   - [ ] Create /import command handler
   - [ ] Add JSON validation
   - [ ] Implement database import logic
   - [ ] Add progress reporting
   - [ ] Add error handling
   - [ ] Test with sample backup

2. Render Configuration
   - [ ] Add health check endpoint
   - [ ] Create render.yaml configuration
   - [ ] Set up environment variables
   - [ ] Configure build process
   - [ ] Set up logging

3. Deployment Steps
   - [ ] Verify backup is current
   - [ ] Configure Render environment
   - [ ] Deploy new version
   - [ ] Verify health check
   - [ ] Test all features
   - [ ] Import data if needed

### Deployment Strategy

1. Pre-Deployment:
   ```bash
   # Verify we have latest backup
   /export
   
   # Create render.yaml
   services:
   - type: web
     name: onmuymind-bot
     env: go
     buildCommand: go build -o mindbot
     startCommand: ./mindbot
     healthCheckPath: /health
     envVars:
     - key: BOT_TOKEN
       sync: false
     - key: OPENROUTER_KEY
       sync: false
     - key: DATA_DIR
       value: /data
     disk:
       name: data
       mountPath: /data
       sizeGB: 1
   ```

2. Deployment Process:
   a. Stage 1: Deploy with new code but keep existing database
   b. Stage 2: If database issues occur, use import feature
   c. Stage 3: Verify all functionality

3. Rollback Plan:
   - Keep old deployment as backup
   - Use exported data for recovery if needed
   - Document all environment settings

## Current Status / Progress Tracking
Phase 1: Completed ✓
Phase 2: Completed ✓
Phase 3: Planned
Phase 4: Planning Stage - Deployment & Data Migration

## Executor's Feedback or Assistance Requests
Ready to begin implementation of Import feature before deployment.

## Example Queries to Support
1. Basic Format Conversion:
   - "What time is 14:00?"
   - "Convert 3:30 PM to 24h format"

2. Time Zone Conversions:
   - "What time is it in Beijing?"
   - "If it's 2pm in Austin, what time is it in Tokyo?"

3. Travel Planning:
   - "If my flight boards at 7:45 AM, security takes 25 minutes, drive takes 1 hour, and breakfast takes 40 minutes, when should I leave?"
   - "Working backwards from a 9 AM meeting, need 30 min commute, 20 min coffee stop, when to leave?"

4. Time Arithmetic:
   - "What's 45 minutes before 7 PM?"
   - "Add 2.5 hours to 3:15 PM"
   - "Subtract 25 minutes from 14:00"

## Lessons
- Environment variables should be properly protected and not committed to repository
- CI checks should verify both build and test success 