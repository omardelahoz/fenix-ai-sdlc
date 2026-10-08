# Good Example

Input

Story:
Design the tables for user authentication and roles.

Expected Result
- Users table with UUID primary key
- Roles table
- UserRoles many-to-many junction table
- Unique constraint on Email
- Initial migration script (UP and DOWN)

---

# Bad Example

- Missing DOWN migration
- Storing passwords in plain text
- Missing foreign key between UserRoles and Users
- Missing indexes on frequently searched columns (e.g. Email)
