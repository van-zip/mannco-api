# Mannco.store API Wrapper

A Go wrapper for the Mannco.store API (`https://docs.mannco.store`)

## Implementation Progress
- [x] **Base Architecture**
  - [x] Global Base URL Configuration
  - [x] Unified API response payload wrapper (`ApiResponse[T]`)
  - [x] Generic HTTP execution pipeline (`executeRequest[T]`)
  - [x] Context (`context.Context`) boundary passing
  - [x] Automatic Bearer Token injection headers
  - [x] Client-level active JWT state management (`SetJWT`)
---
### [Authentication & Session](https://docs.mannco.store/authentication)
- [x] **Login & Tokens**
  - [x] `POST /user/login`  Exchange API key for a fresh session JWT (`UserLogin`)
---
### [Items & Pricing](https://docs.mannco.store/items)
- [x] **Item Discovery & Stats**
  - [x] `GET /item/salesGraph/{item}`  High-resolution historical trade data (`ItemSalesGraph`)
  - [x] `GET /item/listing/{item}`  Paginated market listings with user/bot filtering (`ItemListings`)
  - [x] `GET /item/buyorderList/{item}`  Active multi-tier purchasing demands (`BuyOrderList`)
- [x] **Pricing Evaluations**
  - [x] `GET /item/pricing/{item}`  Individual cached & suggested valuation statistics (`ItemPricing`)
  - [x] `GET /item/pricing/bulk`  Max 100 ID vectorized pricing requests (`ItemPricingBulk`)
- [x] **Market Orders**
  - [x] `POST /item/buyorder`  Create new automatic buy orders (`CreateBuyOrder`)
---
### [User Profiles & History](https://docs.mannco.store/user)
- [x] **Account Data**
  - [x] `GET /user/balance`  Direct wallet integer cents querying (`Balance`)
- [x] **Historical Ledgers**
  - [x] `GET /user/getTransactionHistory`  Internal balances logs (`TransactionHistory`)
  - [x] `GET /user/getSalesHistory`  Searchable vendor transaction lines (`SalesHistory`)
  - [x] `GET /user/getPurchaseHistory`  Historical site acquisitions (`PurchaseHistory`)
  - [ ] `GET /user/getCashoutHistory`
  - [ ] `GET /user/getTransactionDetails`
---
### [Trade Offers](https://docs.mannco.store/offers)
- [ ] **Offer Intake & Operations**
  - [ ] `GET /offers/received`
  - [ ] `GET /offers/my`
  - [ ] `POST /offers/create`
  - [ ] `POST /offers/accept`
  - [ ] `POST /offers
