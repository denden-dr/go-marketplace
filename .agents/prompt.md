# GOALS
- I want to create integration test on handler level, mimic the api call through endpoint
- Integration test should cover all happy path and error path without third party services first, required mocking them or need manual action toward the database
- Integration test should use test container, and run fine, if there are problem with handler/service/repo/model, you should create issue.md with explanation and continue to write test.

# Description


**NOTES:**
* Create the spec
* Create the plan with high level human instruction with no code generation, the plan including:
    - prerequisites
    - add/modify files list
    - edge cases and common problems
    - create each story of how it handle happy/error path
    - update test case from unit test and integration test
put in .agents/plans folder, create if not exist


**MUST**
Read GEMINI.md | CLAUDE.md 
if exist