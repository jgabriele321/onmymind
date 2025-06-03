# Project Scratchpad

## Project Phases
1. ✓ Initial Setup and Configuration (Completed)
2. ✓ Time Calculator AI Assistant (Completed)
3. Enhanced Reminder System (Planned)
4. Airport Travel Time Feature (Planned)
5. Render Deployment & Data Migration
6. Local Development Environment Setup (New)

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

## Enhanced Reminder System Implementation Plan (Phase 3)

### Day 1: Core Infrastructure Setup

1. Database Schema Design (2 hours)
   Success Criteria:
   - Schema supports both one-off and recurring reminders
   - Stores reminder context and metadata
   - Handles time zones correctly
   - Supports multiple notification methods
   Implementation Steps:
   a. Create reminders table with fields:
      - id (UUID)
      - user_id (string)
      - title (string)
      - description (text)
      - due_time (timestamp with timezone)
      - recurrence_pattern (string, nullable)
      - priority (boolean, default false)  // Indicates if this is a priority reminder
      - status (enum: pending, completed, cancelled)
      - created_at (timestamp)
      - updated_at (timestamp)
   b. Create reminder_logs table for tracking notifications:
      - id (UUID)
      - reminder_id (UUID)
      - notification_type (enum: whatsapp_message, whatsapp_call)
      - status (enum: success, failed)
      - error_message (text, nullable)
      - attempted_at (timestamp)

2. Command Parser Development (2 hours)
   Success Criteria:
   - Parses natural language time inputs
   - Handles recurring patterns
   - Validates user input
   - Provides helpful error messages
   Implementation Steps:
   a. Create command handlers:
      - /remindme <time> <message> [-call]  // Regular reminder, -call flag for priority
      - /reminders [all|priority|regular]  // List reminders with optional filter
      - /delete <reminder_id>  // Delete a specific reminder
      - /complete <reminder_id>  // Mark a reminder as completed
   b. Implement time parser for formats:
      - Relative time: "in 2 hours", "tomorrow at 3pm"
      - Absolute time: "2024-03-20 15:00"
      - Natural recurring patterns:
        - "every Sunday at 10am"
        - "every day at 2pm"
        - "every month on the first"
        - "every Monday and Wednesday at 3pm"
        - "every weekday at 9am"

### Day 2: Notification System

1. Platform Integration (3 hours)
   Success Criteria:
   - Successfully integrates with WhatsApp messages and calls
   - Handles rate limits appropriately
   - Implements retry mechanism
   - Logs all communication attempts
   Implementation Steps:
   a. Create NotificationService interface
   b. Implement WhatsAppMessageService
   c. Implement WhatsAppCallService
   d. Create NotificationOrchestrator for priority reminders:
      - Send WhatsApp message first
      - If no acknowledgment within 2 minutes, initiate WhatsApp call
   e. Add rate limiting and retry logic
   f. Set up error handling and logging
   g. Create mock services for testing

2. Reminder Execution System (3 hours)
   Success Criteria:
   - Accurately triggers reminders at specified times
   - Handles recurring reminder rescheduling
   - Manages notification priorities correctly
   - Provides status updates
   Implementation Steps:
   a. Create ReminderExecutor service
   b. Implement scheduling system
   c. Add recurrence handler
   d. Create notification queue with priority handling:
      - Regular reminders: Single WhatsApp message
      - Priority reminders: WhatsApp message + conditional call
   e. Set up monitoring and alerts

### Day 3: Testing and Polish

1. Testing Suite Development (3 hours)
   Success Criteria:
   - >80% test coverage
   - All core flows tested
   - Edge cases handled
   - Performance tests pass
   Implementation Steps:
   a. Write unit tests for:
      - Command parsing
      - Time calculations
      - Notification system
      - Database operations
   b. Create integration tests
   c. Add performance benchmarks
   d. Test edge cases:
      - Timezone changes
      - Daylight saving transitions
      - Network failures
      - Rate limit handling

2. User Experience Enhancement (2 hours)
   Success Criteria:
   - Clear error messages
   - Helpful command examples
   - Status feedback for all operations
   - Easy reminder management
   Implementation Steps:
   a. Add detailed help messages
   b. Implement progress indicators
   c. Create user-friendly responses
   d. Add confirmation messages

3. Documentation and Deployment Prep (1 hour)
   Success Criteria:
   - Code is well documented
   - README is updated
   - Deployment steps are clear
   Implementation Steps:
   a. Update code documentation
   b. Create deployment guide
   c. Document configuration options
   d. Add troubleshooting guide

### Project Status Board for Phase 3

Day 1:
- [ ] Database Schema Implementation
  - [ ] Create reminders table
  - [ ] Create reminder_logs table
  - [ ] Add indexes and constraints
  - [ ] Write migrations
- [ ] Command Parser Development
  - [ ] Implement basic commands
  - [ ] Add time parsing
  - [ ] Create input validation
  - [ ] Add error handling

Day 2:
- [ ] Platform Integration
  - [ ] Set up WhatsApp message integration
  - [ ] Set up WhatsApp call integration
  - [ ] Implement rate limiting
  - [ ] Add retry mechanism
  - [ ] Create logging system
- [ ] Reminder Execution System
  - [ ] Create scheduler
  - [ ] Add recurrence handling
  - [ ] Implement notification queue
  - [ ] Set up monitoring

Day 3:
- [ ] Testing Suite
  - [ ] Write unit tests
  - [ ] Create integration tests
  - [ ] Add performance tests
  - [ ] Test edge cases
- [ ] User Experience
  - [ ] Improve error messages
  - [ ] Add progress indicators
  - [ ] Create help documentation
- [ ] Documentation
  - [ ] Update README
  - [ ] Create deployment guide
  - [ ] Document configuration

### Success Metrics
1. Functional Requirements:
   - All reminders trigger within ±1 minute of scheduled time
   - Recurring reminders properly reschedule
   - Failed notifications retry up to 3 times
   - All reminder operations (create, delete, complete) work reliably

2. Performance Requirements:
   - Command response time < 2 seconds
   - Notification triggering latency < 5 seconds
   - System handles 100+ concurrent reminders
   - Database queries complete < 100ms

3. User Experience Requirements:
   - Clear feedback for all operations
   - Helpful error messages
   - Easy to manage reminders
   - Intuitive command syntax

### Risk Assessment and Mitigation

1. Technical Risks:
   - Rate limiting from WhatsApp
     Mitigation: Implement smart queuing and fallback mechanisms
   - Database performance
     Mitigation: Proper indexing and query optimization
   - Time zone handling
     Mitigation: Comprehensive testing across time zones

2. User Experience Risks:
   - Complex command syntax
     Mitigation: Natural language processing and examples
   - Notification reliability
     Mitigation: Multiple retry attempts and status updates

### Questions for Tomorrow
1. Should priority reminders have a cooldown period between message and call?
2. What should be the maximum number of priority reminders allowed per day?
3. How should we handle acknowledgments to prevent unnecessary calls?
4. What metrics should we track for monitoring both regular and priority notifications?
5. Should users be able to upgrade a regular reminder to priority or vice versa?

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

### Local Development Tasks
- [ ] Create TestMain.go
  - [ ] Environment variable loading
  - [ ] Development configuration
  - [ ] Debug logging
  - [ ] Hot reload support
- [ ] Environment Setup
  - [ ] Create .env.example
  - [ ] Document variables
  - [ ] Implement validation
- [ ] Development Tools
  - [ ] Install air
  - [ ] Configure hot reload
  - [ ] Create dev scripts
- [ ] Documentation
  - [ ] Update README
  - [ ] Add development guide

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

Need clarification on:
1. Should TestMain.go support both bot modes (CLI and Telegram)?
2. What development-specific features are most important?
3. Should we implement mock services for testing?

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

5. Reminder Examples:
   - One-time Reminders:
     - "/remindme tomorrow at 3pm to call mom"
     - "/remindme in 2 hours to check email"
     - "/remindme 2024-03-20 15:00 to submit report -call"
   - Recurring Reminders:
     - "/remindme every Sunday at 10am to water plants"
     - "/remindme every day at 2pm to take a break"
     - "/remindme every month on the first to pay rent -call"
     - "/remindme every Monday and Wednesday at 3pm to attend team meeting"
     - "/remindme every weekday at 9am to check emails"
   - Managing Reminders:
     - "/reminders" - List all reminders
     - "/reminders priority" - List priority (call-enabled) reminders
     - "/reminders regular" - List regular reminders
     - "/delete 123" - Delete reminder with ID 123
     - "/complete 456" - Mark reminder 456 as completed

## Lessons
- Environment variables should be properly protected and not committed to repository
- CI checks should verify both build and test success
- Always provide clear error messages for missing configuration
- Use environment variables for sensitive data
- Keep development and production code paths separate
- Document setup steps clearly

# Time Zone Implementation Analysis and Plan

## Background and Motivation
The bot needs to handle time zone calculations accurately and reliably. While we have IANA time zone database integration, we need a more robust way to ensure the LLM uses it correctly. Instead of relying on prompt engineering, we can create dedicated tool functions that the LLM can call directly.

## Key Challenges and Analysis

1. **Current Implementation Issues**
   - The LLM sometimes ignores the provided time zone data and does manual calculations
   - Complex prompt engineering is fragile and hard to maintain
   - Limited city-to-timezone mapping
   - No validation or fuzzy matching for city names

2. **Proposed Tool Function Approach**
   - Create IANA tool functions that the LLM can call directly
   - Move time zone logic out of the prompt and into Go code
   - Provide structured responses that the LLM can easily parse and format
   - Handle all time zone calculations in Go using the IANA database

3. **Benefits of Tool Function Approach**
   - Forces the LLM to use IANA data (can't do manual calculations)
   - Centralizes time zone logic in Go code where it belongs
   - Easier to test and maintain
   - More reliable and accurate results
   - Better error handling and validation

## High-level Task Breakdown

1. **Design IANA Tool Functions** [⏳ Not Started]
   - Create `GetCurrentTime(location string) TimeInfo`
   - Create `ConvertTime(time string, fromZone string, toZone string) TimeConversion`
   - Create `GetTimeZoneInfo(location string) ZoneInfo`
   - Create `ValidateLocation(location string) LocationValidation`
   Success Criteria: Complete set of tool functions that handle all time zone operations

2. **Implement Time Zone Tools** [⏳ Not Started]
   - Implement each tool function in Go
   - Use IANA database directly for all calculations
   - Add comprehensive error handling
   - Add input validation and normalization
   - Add fuzzy matching for location names
   Success Criteria: Tool functions work correctly for all test cases

3. **Update OpenRouter Integration** [⏳ Not Started]
   - Add tool function definitions to system prompt
   - Update example queries to show tool usage
   - Remove time zone calculation logic from prompt
   - Add error handling guidelines
   Success Criteria: LLM consistently uses tool functions for all time zone operations

4. **Enhance Location Handling** [⏳ Not Started]
   - Create location database with common names
   - Add country information for disambiguation
   - Implement fuzzy matching
   - Handle abbreviations (EST, GMT, etc.)
   Success Criteria: System can handle a wide range of location inputs

5. **Implement Testing Suite** [⏳ Not Started]
   - Unit tests for all tool functions
   - Integration tests with OpenRouter
   - Test edge cases (DST, invalid locations)
   - Test different query formats
   Success Criteria: >90% test coverage and all edge cases handled

## Revised Example System Prompt for Time Zone Tools

You are a time zone assistant with access to the following tool functions:

1. GetCurrentTime(location)
   Input: City or location name
   Returns: Current time, zone name, and DST status
   Example: GetCurrentTime("New York") → "2:30 PM EDT (UTC-4), DST in effect"

2. ConvertTime(time, fromZone, toZone)
   Input: Time expression and location names
   Returns: Converted time with zone details
   Example: ConvertTime("2:30 PM", "New York", "Tokyo") → "3:30 AM JST next day"

3. GetTimeZoneInfo(location)
   Input: City or location name
   Returns: Zone name, offset, and DST information
   Example: GetTimeZoneInfo("London") → "Europe/London, UTC+0, DST starts March 31"

4. ValidateLocation(location)
   Input: City or location name
   Returns: Validation result and suggestions
   Example: ValidateLocation("NYC") → "Valid: New York City, United States"

IMPORTANT RULES:
1. ALWAYS use these tools for time zone operations
2. NEVER perform manual time calculations
3. NEVER assume time zones or offsets
4. Validate locations before using them
5. Show both 12h and 24h time formats
6. Include DST information when relevant

Example Usage:

Q: "What time is it in Tokyo?"
A: Let me check the current time in Tokyo.
First, I'll validate the location:
Tool: ValidateLocation("Tokyo")
Now I'll get the current time:
Tool: GetCurrentTime("Tokyo")
Response: "In Tokyo, it's 3:30 PM (15:30) JST. Japan does not observe daylight saving time."

Q: "If it's 2pm in New York, what time is it in London?"
A: I'll help you with that conversion.
1. Validate both locations:
   Tool: ValidateLocation("New York")
   Tool: ValidateLocation("London")
2. Convert the time:
   Tool: ConvertTime("2:00 PM", "New York", "London")
Response: "When it's 2:00 PM (14:00) in New York, it's 7:00 PM (19:00) in London"

Q: "What's the time difference between Paris and Sydney?"
A: Let me check both time zones.
1. Get information for both cities:
   Tool: GetTimeZoneInfo("Paris")
   Tool: GetTimeZoneInfo("Sydney")
Response: "Paris (UTC+1) and Sydney (UTC+10) are 9 hours apart. Paris observes DST from March to October, while Sydney observes DST from October to April."

## Project Status Board
- [ ] Task 1: Design IANA Tool Functions
- [ ] Task 2: Implement Time Zone Tools
- [ ] Task 3: Update OpenRouter Integration
- [ ] Task 4: Enhance Location Handling
- [ ] Task 5: Implement Testing Suite

## Executor's Feedback or Assistance Requests
[No feedback yet]

## Lessons
- Time zone calculations should be handled by dedicated tool functions
- Keep time zone logic in Go code, not in prompts
- Use structured types for clear data exchange with LLM
- Always validate and normalize location inputs 

### Local Development Setup Analysis
1. Technical Requirements:
   - Separate configuration for local development
   - Environment variable management
   - Easy-to-use startup process
   - Development-specific features (debug logging, etc.)
   - Hot reload capability for faster development

2. User Experience Considerations:
   - Simple command to start the bot locally
   - Clear error messages for missing configuration
   - Easy switching between development and production
   - Helpful debugging output

3. Potential Challenges:
   - Managing different configurations between environments
   - Keeping production and development code in sync
   - Handling sensitive data in development
   - Testing Telegram bot features locally

### Local Development Implementation Plan (Non-invasive Approach)
1. Create Parallel Development Environment
   - Success Criteria:
     - Development environment runs alongside production code
     - No changes to existing production files
     - No impact on Render deployment
   - Implementation Steps:
     a. Create `dev/` directory for development files
     b. Create `dev/main.go` for local testing
     c. Create `dev/.env` for development variables
     d. Add `dev/` to .gitignore

2. Environment Configuration
   - Success Criteria:
     - Development uses separate .env file
     - No interference with production environment
     - Clear documentation of required variables
   - Implementation Steps:
     a. Create `dev/.env.example` template
     b. Document development-specific variables
     c. Keep production .env separate and unchanged

3. Development Tools
   - Success Criteria:
     - Local testing environment works independently
     - Easy to switch between dev and prod
     - Clear separation of concerns
   - Implementation Steps:
     a. Create development-specific make targets
     b. Add debug logging for development
     c. Document local development process

4. Documentation
   - Success Criteria:
     - Clear instructions for local development
     - No confusion between dev and prod environments
     - Easy onboarding process
   - Implementation Steps:
     a. Add development guide to README
     b. Document environment differences
     c. Provide example usage

### Updated Task List
- [ ] Development Environment Setup
  - [ ] Create `dev/` directory
  - [ ] Create `dev/main.go`
  - [ ] Create `dev/.env.example`
  - [ ] Update .gitignore
- [ ] Documentation
  - [ ] Add development setup guide
  - [ ] Document environment differences
  - [ ] Add example usage

### Key Principles
1. No changes to production code
2. Keep Render deployment stable
3. Maintain clear separation between environments
4. Document everything clearly 

## Airport Travel Time Feature Analysis (Phase 4)

### Background and Motivation
Users want to optimize their travel time to the airport by considering historical and real-time traffic data. This would help them decide the best time to leave for their flight.

### Key Challenges and Analysis

1. Data Requirements:
   - Real-time traffic data
   - Historical traffic patterns
   - Route information
   - Airport-specific data

2. Potential APIs:
   - Google Maps API
     - Pros: Most comprehensive traffic data, reliable routing
     - Cons: Usage costs, rate limits
   - HERE Maps API
     - Pros: Good traffic coverage, competitive pricing
     - Cons: Less detailed than Google in some areas
   - TomTom API
     - Pros: Specialized in traffic data, good historical patterns
     - Cons: Coverage might vary by region
   - Waze API
     - Pros: Real-time user-reported incidents
     - Cons: Limited historical data

3. Additional Data Points Needed:
   - Flight information (departure time)
   - Airport security wait times
     - Could use TSA API for US airports
     - Third-party APIs like FlightStats
   - Terminal walking times
   - Check-in/baggage drop requirements

### Implementation Plan

1. API Selection and Integration
   - Success Criteria:
     - Selected API provides reliable coverage for target airports
     - Response times under 2 seconds
     - Cost-effective for expected usage
     - Comprehensive traffic data available
   Implementation Steps:
     a. Research and compare API pricing
     b. Test API response times
     c. Verify data quality
     d. Implement selected API

2. Command Design
   - Success Criteria:
     - Clear, user-friendly command syntax
     - All necessary parameters captured
     - Helpful error messages
     - Example usage provided
   Implementation Steps:
     a. Design `/airport` command structure
     b. Define required parameters
     c. Create help documentation
     d. Add input validation

3. Core Functionality
   - Success Criteria:
     - Accurate route calculations
     - Real-time traffic integration
     - Historical pattern analysis
     - Buffer time recommendations
   Implementation Steps:
     a. Implement route calculation
     b. Add traffic analysis
     c. Create buffer time algorithm
     d. Add time recommendations

4. Testing Suite
   - Success Criteria:
     - Predictions within 15-minute accuracy
     - Reliable across different times/days
     - Handles edge cases properly
   Implementation Steps:
     a. Create unit tests
     b. Add integration tests
     c. Implement stress testing
     d. Document test cases

### Task Breakdown

1. Initial Setup
   - [ ] Research and select traffic API
   - [ ] Set up API authentication
   - [ ] Create configuration structure
   - [ ] Add API documentation

2. Command Implementation
   - [ ] Create `/airport` command handler
   - [ ] Add parameter parsing
   - [ ] Implement input validation
   - [ ] Add help text and examples

3. Core Features
   - [ ] Build route calculator
   - [ ] Add traffic analysis
   - [ ] Implement buffer calculator
   - [ ] Create time optimizer

4. Testing and Validation
   - [ ] Write unit tests
   - [ ] Create integration tests
   - [ ] Add performance tests
   - [ ] Document test scenarios

### Example Usage

```
/airport JFK "2024-04-01 14:30" "123 Main St, Brooklyn"
Response: Based on your 2:30 PM flight from JFK:
- Estimated drive time: 45-60 minutes
- Current traffic: Moderate
- Security wait: ~15 minutes
- Recommended departure: 12:30 PM
- Buffer included: 45 minutes
```

### Questions for User
1. Which airports do you primarily travel from/to?
2. Do you have a preference for any of the mentioned APIs?
3. What's the acceptable margin of error for time predictions?
4. Would you want flight status integration as well?

### Lessons
- Keep track of API rate limits and costs
- Consider caching historical data to reduce API calls
- Plan for API fallbacks in case of service disruptions 