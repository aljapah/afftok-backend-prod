# Integration Testing Checklist

## Backend Testing

### 1. API Endpoint Tests
- [ ] Test `/api/promoter/user/{username}` with valid username
  - Expected: Returns HTML page with user profile and offers
  - Status Code: 200
  
- [ ] Test `/api/promoter/user/{username}` with invalid username
  - Expected: Returns 404 error
  - Status Code: 404

- [ ] Test `/api/promoter/{id}` with valid UUID (backward compatibility)
  - Expected: Still works as before
  - Status Code: 200

- [ ] Test static file serving
  - [ ] `/public/promoter_landing.html` - Should return HTML
  - [ ] `/public/privacy.html` - Should return HTML
  - [ ] `/public/terms.html` - Should return HTML

### 2. Database Verification
- [ ] Verify user exists in database with username
- [ ] Verify offers are marked as "active" in database
- [ ] Verify user_offers relationship is correct

### 3. HTML Generation
- [ ] Verify HTML contains user's full name
- [ ] Verify HTML contains user's bio
- [ ] Verify HTML contains all active offers
- [ ] Verify offer cards display correctly
- [ ] Verify bilingual content is present

## Flutter App Testing

### 1. Profile Screen
- [ ] Navigate to Profile screen
- [ ] Verify "صفحتي" button is visible
- [ ] Click "صفحتي" button
- [ ] Verify WebView opens with correct URL

### 2. Landing Page Display
- [ ] Verify page loads in WebView
- [ ] Verify user profile is displayed
- [ ] Verify offers are displayed
- [ ] Verify responsive layout on different screen sizes
- [ ] Verify bilingual toggle works

### 3. User Interactions
- [ ] Click on offer cards
- [ ] Click social media links
- [ ] Click app download buttons
- [ ] Verify language toggle switches between Arabic/English

## Device Testing

### Desktop (1920x1080)
- [ ] 3 offer cards per row
- [ ] All elements properly aligned
- [ ] Responsive design working

### Tablet (768x1024)
- [ ] 2 offer cards per row
- [ ] Touch interactions working
- [ ] Layout properly adjusted

### Mobile (375x667)
- [ ] 1 offer card per row
- [ ] Scrolling works smoothly
- [ ] Touch interactions responsive

## Bilingual Testing

### English
- [ ] Page title displays in English
- [ ] All buttons display in English
- [ ] Navigation works correctly

### Arabic
- [ ] Page title displays in Arabic
- [ ] All buttons display in Arabic
- [ ] RTL layout is correct
- [ ] Navigation works correctly

## Performance Testing

- [ ] Page loads within 2 seconds
- [ ] Images load properly
- [ ] No console errors
- [ ] No memory leaks in WebView

## Security Testing

- [ ] SQL injection attempts fail
- [ ] XSS attempts are blocked
- [ ] CORS headers are correct
- [ ] User data is properly escaped in HTML

## Edge Cases

- [ ] User with no offers
- [ ] User with many offers (50+)
- [ ] User with special characters in username
- [ ] User with very long bio
- [ ] Offers with missing images
- [ ] Offers with very long descriptions

## Production Deployment

- [ ] Code changes pushed to Git
- [ ] Railway deployment successful
- [ ] All endpoints accessible
- [ ] Static files served correctly
- [ ] No errors in production logs

## Sign-Off

- [ ] All tests passed
- [ ] No breaking changes
- [ ] Backward compatibility maintained
- [ ] Ready for production

---

**Test Date**: _______________
**Tested By**: _______________
**Status**: _______________
