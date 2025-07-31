# User Registration Flow

This diagram illustrates the process for a new user registering for an account, from initial sign-up to account activation.

```mermaid
flowchart TD
    A[Start] --> B(User navigates to registration page);
    B --> C{Fills out registration form<br/>(Name, Email, Password)};
    C --> D[System validates input];
    D --> E{Is data valid?};
    E -- No --> F[Display validation errors on form];
    F --> C;
    E -- Yes --> G[Create user account with 'Pending' status];
    G --> H[Generate & send verification email];
    H --> I(User receives email and clicks verification link);
    I --> J[System verifies token from link];
    J --> K{Is token valid?};
    K -- No --> L[Display error page: "Invalid or expired link"];
    K -- Yes --> M[Update user status to 'Active'];
    M --> N[Redirect to login page with success message];
    N --> O[End];
```
