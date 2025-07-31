# ADR 001: Use JWT for Authentication

## Status

Accepted

## Context

We need a secure and scalable way to authenticate users for our API.

## Decision

We will use JSON Web Tokens (JWT) for authentication.

## Consequences

- Stateless authentication, which is good for scalability.
- Frontend will be responsible for storing the JWT.
- We will need to implement a token refresh mechanism.
