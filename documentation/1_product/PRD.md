# Product Requirements Document: BookingNow OTA Platform

## 1. Overview & Vision
BookingNow is a comprehensive Online Travel Agency (OTA) for the Singapore market, designed to be a one-stop-shop for booking transportation, accommodation, and activities. Our vision is to simplify travel planning by providing a unified, seamless, and mobile-first user experience.

## 2. Target Audience & Personas
Our primary target audience consists of tech-savvy residents and visitors in Singapore looking for convenient ways to plan and book local and regional travel.

**Persona 1: "Sarah the Weekend Explorer"**
| Attribute | Description |
| :--- | :--- |
| **Age** | 28 |
| **Occupation** | Marketing Manager |
| **Behavior** | Plans short weekend trips from Singapore to nearby islands (Batam, Bintan) once a quarter. |
| **Needs** | A fast, mobile-friendly way to book a ferry and a hotel in one go. Values clear pricing and user reviews. |
| **Frustrations** | Having to visit multiple websites to book different parts of her trip. |

**Persona 2: "David the Family Planner"**
| Attribute | Description |
| :--- | :--- |
| **Age** | 42 |
| **Occupation** | IT Consultant |
| **Behavior** | Organizes family holidays for himself, his wife, and two children (ages 8 and 12) twice a year. |
| **Needs** | A platform that can handle bookings for multiple people, including different ticket types (adult/child). Wants package deals for convenience and value. |
| **Frustrations** | Difficulty coordinating travel times and accommodation availability for a group. |

## 3. Business Objectives & Success Metrics

**Objectives:**
- Launch the platform and acquire the first 1,000 paying customers within 6 months.
- Onboard at least 10 ferry operators and 50 hotels by the end of Year 1.
- Achieve a customer satisfaction (CSAT) score of 85% or higher.

**Key Performance Indicators (KPIs):**
- Gross Merchandise Value (GMV) > $100K in the first year.
- Customer Conversion Rate > 3%.
- Average Booking Value > $150.
- Customer Acquisition Cost (CAC) < $15.

## 4. Key Features & Scope

### User Management (Core)
- **User Registration:** Standard email/password sign-up and social login (Google, Facebook).
- **User Profile:** Manage personal details, view booking history, and save payment methods.
- **Role-Based Access Control (RBAC):**
    - **Customer:** Standard user who can search and book.
    - **Operator:** Partner (e.g., hotel manager) who can manage listings and view bookings.
    - **Admin:** Internal staff for platform management and support.

### Transport Ticketing (Phase 1)
- **Ferry Booking Engine:**
    - Search for ferry routes based on origin, destination, and date.
    - View available operators, schedules, and prices.
    - Select seats and add passengers.
- **Bus Booking Engine (Phase 2):**
    - Search for bus routes.
    - View schedules and amenities.

### Accommodation Booking (Phase 1)
- **Hotel Search Engine:**
    - Search for hotels by destination, dates, and number of guests.
    - Filter results by price, star rating, and amenities.
    - View hotel details, photos, and user reviews.
- **Room Booking:**
    - Select room types and view availability.
    - Add rooms to the shopping cart.

### Unified Booking & Payments (Core)
- **Centralized Shopping Cart:** Add multiple items (e.g., ferry tickets and a hotel room) to a single cart before checkout.
- **Payment Gateway Integration:** Secure payment processing via Stripe, supporting major credit cards.
- **Booking Confirmation:** Automated email confirmation with e-tickets and booking vouchers.

## 5. Assumptions & Constraints

**Assumptions:**
- Users are comfortable with digital payments and e-tickets.
- Ferry and hotel operators are willing to provide API access or manually manage their listings for a commission-based fee.
- The target audience is primarily mobile-first.

**Constraints:**
- The initial launch (Phase 1) will only include ferry and hotel bookings for routes originating from Singapore.
- Bus and activity bookings will be deferred to Phase 2 to ensure a focused and stable initial launch.
- The platform must comply with Singapore's Personal Data Protection Act (PDPA).

## 6. Release Plan / Phasing

This roadmap outlines the planned phases for the BookingNow platform launch and subsequent enhancements.

**(Text-based representation of the roadmap)**

**Phase 1: Core Ferry & Hotel Booking Launch (Target: Q4 2025)**
- **Features:**
    - User Registration & Profile Management
    - Ferry Search & Booking
    - Hotel Search & Booking
    - Unified Shopping Cart & Stripe Payments
- **Goal:** Establish a minimum viable product (MVP) and validate the core business model.

**Phase 2: Expansion of Services (Target: Q1 2026)**
- **Features:**
    - Bus Booking Engine
    - Attraction & Tour Ticket Booking
    - User Reviews & Ratings System
- **Goal:** Increase the platform's value proposition and capture a larger share of the travel wallet.

**Phase 3: Enhancement & Personalization (Target: Q2 2026)**
- **Features:**
    - Package Deals (Ferry + Hotel bundles)
    - User-based recommendations
    - Loyalty Program
- **Goal:** Improve user retention and increase average booking value.
