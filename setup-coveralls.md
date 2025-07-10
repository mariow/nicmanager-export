# Coveralls Setup Instructions

## Quick Setup Steps

### 1. Enable Repository on Coveralls.io

1. **Visit Coveralls**: Go to https://coveralls.io/
2. **Sign In**: Click "Sign In" and authorize with your GitHub account
3. **Add Repository**: 
   - Go to https://coveralls.io/repos/new
   - Find `mariow/nicmanager-export` in the list
   - Toggle the switch to **ON** to enable coverage tracking
4. **Get Token** (if needed):
   - Click "START UPLOADING COVERAGE" next to your repo
   - Note the repo token (usually not needed for public repos)

### 2. Verify Integration

The GitHub Actions workflow has been updated to automatically upload coverage data. After the next push to master:

1. **Check Actions**: Go to https://github.com/mariow/nicmanager-export/actions
2. **Verify Upload**: Look for "Upload coverage to Coveralls" step in the workflow
3. **Check Coveralls**: Visit https://coveralls.io/github/mariow/nicmanager-export
4. **Badge Update**: The README badge should show actual coverage percentage

### 3. Test with Pull Request

Create a test PR to verify:
- Coverage data is uploaded
- Coveralls adds a comment showing coverage changes
- Badge reflects current coverage

## Troubleshooting

### If the badge still shows "unknown":
1. Ensure the repository is enabled on Coveralls.io
2. Check that at least one successful workflow run has completed
3. Wait a few minutes for Coveralls to process the data

### If you get "Repository not found" errors:
1. Double-check the repository is toggled ON at https://coveralls.io/repos/new
2. Verify the repository name matches exactly: `mariow/nicmanager-export`

### For private repositories:
1. Get the repo token from Coveralls.io
2. Add it as a GitHub secret named `COVERALLS_REPO_TOKEN`
3. Update the workflow to use the token instead of `GITHUB_TOKEN`

## Current Status

✅ **GitHub Actions workflow updated** with proper Coveralls integration  
✅ **Documentation created** with setup instructions  
⏳ **Repository needs to be enabled** on Coveralls.io (manual step)  
⏳ **First coverage upload** will happen on next push to master  

The technical integration is complete - only the Coveralls.io repository enablement remains.