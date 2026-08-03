-- Seeds the Help Center FAQ list (help_faqs table, managed from
-- admin-portal's /dashboard/support page, served publicly via
-- GET /api/help to seller-app/customer-app). Fixed ids so this migration is
-- idempotent (ON CONFLICT DO NOTHING) and admins can still freely edit or
-- deactivate any of these afterward without this migration re-inserting them.
INSERT INTO help_faqs (id, question, answer, sort_order, is_active) VALUES
('faq_seed_0001', 'How do I create my shop on Locazar?', 'Download the seller app, verify your phone number with an OTP, then complete the shop setup wizard: shop name, address, store category, and a shop photo.', 1, TRUE),
('faq_seed_0002', 'Why do I need to verify my phone number twice — once at signup and once for my shop?', 'The signup OTP confirms it''s really you; some shop actions (like changing your registered phone later) need a second OTP to prove you still own that number.', 2, TRUE),
('faq_seed_0003', 'What is a Referral ID, and do I need one?', 'It''s optional. If a Locazar Sales Executive gave you a code during signup, entering it links your shop to them. Leaving it blank doesn''t affect your onboarding at all.', 3, TRUE),
('faq_seed_0004', 'I entered a Referral ID but it says "No referral found" — what happened?', 'The code doesn''t match any active Sales Executive account. Double-check for typos, or continue without it — your shop is created either way.', 4, TRUE),
('faq_seed_0005', 'How do I get my shop verified?', 'Upload a business document (GSTIN, PAN, Aadhaar, URN, or BRN) and verify it via OTP during onboarding, or later from the Verification section.', 5, TRUE),
('faq_seed_0006', 'Can I skip document verification and add it later?', 'Yes — tap "Skip" during onboarding. Your shop still goes live; a support agent may follow up to complete verification.', 6, TRUE),
('faq_seed_0007', 'What counts as a "fully verified" shop?', 'Your shop photo and address proof are the two required checks. Business and identity documents are optional and don''t block verification.', 7, TRUE),
('faq_seed_0008', 'Why is my phone number visible to customers?', 'During onboarding you''re asked to consent to showing your number so customers can reach you directly. It''s checked by default — you can uncheck it any time from onboarding or your shop profile.', 8, TRUE),
('faq_seed_0009', 'If I hide my number, will customers still be able to contact me?', 'They may find it harder to reach you directly for quick questions before ordering — we show a warning when you turn this off for that reason.', 9, TRUE),
('faq_seed_0010', 'How do I change my shop''s registered phone number?', 'From your shop profile, enter the new number and tap "Get OTP" — the number only updates once you verify it.', 10, TRUE),
('faq_seed_0011', 'How do I add products to my shop?', 'From the Add Products page, pick a category, add photos, fill in details, and publish — no need to wait for full verification first.', 11, TRUE),
('faq_seed_0013', 'I''m a Sales Executive — how do I see shops I''ve referred?', 'From admin-portal''s Platform Users page, your profile shows a "Shops (N)" count and list of shops that used your referral code.', 13, TRUE),
('faq_seed_0014', 'Who do I contact if I''m stuck?', 'Use the in-app Help/Support section, or reach out via the contact details listed in this Help Center.', 14, TRUE)
ON CONFLICT (id) DO NOTHING;
