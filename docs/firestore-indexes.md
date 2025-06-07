# Firestore Index Configuration

This document describes the Firestore composite indexes required for optimal performance of the Newsletter Service.

## Required Indexes

### 1. Subscribers Collection - Active Subscribers Query

**Collection:** `subscribers`
**Query Type:** Composite Index
**Fields:**
- `newsletter_id` (Ascending)
- `status` (Ascending) 
- `subscription_date` (Descending)

**Usage:** Used for listing active subscribers with pagination and ordering by subscription date.

### 2. Subscribers Collection - All Subscribers Query

**Collection:** `subscribers`
**Query Type:** Composite Index  
**Fields:**
- `newsletter_id` (Ascending)
- `subscription_date` (Descending)

**Usage:** Used for listing all subscribers (regardless of status) with pagination and ordering.

## How to Create Indexes

### Option 1: Firebase Console (Recommended)

1. Go to the [Firebase Console](https://console.firebase.google.com/)
2. Select your project: `strv-vse-go-newsletter-jochim`
3. Navigate to **Firestore Database** → **Indexes**
4. Click **"Create Index"**
5. Configure each index with the fields listed above

### Option 2: Automatic Creation via Error Links

When you run queries that require indexes, Firestore will provide direct links to create the required indexes. Look for error messages like:

```
The query requires an index. You can create it here: https://console.firebase.google.com/v1/r/project/strv-vse-go-newsletter-jochim/firestore/indexes?create_composite=...
```

Click these links to automatically create the required indexes.

### Option 3: Firebase CLI

Create a `firestore.indexes.json` file:

```json
{
  "indexes": [
    {
      "collectionGroup": "subscribers",
      "queryScope": "COLLECTION",
      "fields": [
        {
          "fieldPath": "newsletter_id",
          "order": "ASCENDING"
        },
        {
          "fieldPath": "status", 
          "order": "ASCENDING"
        },
        {
          "fieldPath": "subscription_date",
          "order": "DESCENDING"
        }
      ]
    },
    {
      "collectionGroup": "subscribers", 
      "queryScope": "COLLECTION",
      "fields": [
        {
          "fieldPath": "newsletter_id",
          "order": "ASCENDING"
        },
        {
          "fieldPath": "subscription_date",
          "order": "DESCENDING" 
        }
      ]
    }
  ]
}
```

Then deploy with:
```bash
firebase deploy --only firestore:indexes
```

## Current Workaround

The current implementation removes `OrderBy` clauses from complex queries to avoid requiring composite indexes. This allows the application to work immediately but without guaranteed ordering for pagination.

For production use, create the indexes above and uncomment the `OrderBy` clauses in:
- `internal/layers/repository/subscriber.go`
  - `ListSubscribersByNewsletterID()`
  - `ListActiveSubscribersByNewsletterID()`

## Index Build Time

- New indexes can take several minutes to build
- Existing data will be indexed automatically
- The application will work with degraded performance until indexes are complete

## Troubleshooting

If you encounter "index required" errors:
1. Check the error message for the direct creation link
2. Click the link to create the index automatically  
3. Wait for the index to build (usually 1-5 minutes)
4. Retry the operation 