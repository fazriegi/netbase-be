# Fintrack BE

###### Fintrack BE

FinTrack BE is a backend application for managing assets, liabilities, and personal finance tracking.

## Technology Stack

`Go Programming Language` `PostgreSQL`

## Features

- User Registration
- User Login
- Refresh Token

## Database Design

![ERD](db/ERD.png)

## Demo

**LIVE API** : `https://link-to-live-api`  
**API Documentation** : `https://link-to-api-docs`

## Installation

Follow these steps to install and run Fintrack BE on your local machine:

1. **Clone the repository:**

   ```bash
   git clone https://github.com/fazriegi/fintrack-be.git
   ```

2. **Move to cloned repository folder**

   ```bash
   cd fintrack-be
   ```

3. **Update dependecies**

   ```bash
   go mod tidy
   ```

4. **Copy `.env.example` to `.env`**

   ```bash
   cp .env.example .env
   ```

5. **Configure your `.env`**
6. **Migrate the db migrations**
7. **Build and Run the app**

   ```bash
   make run
   ```

## Author

Fazri Egi - [Github](https://github.com/fazriegi)
