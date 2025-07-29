#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Comprehensive analysis of the dot-bypasser extension
function comprehensiveAnalysis() {
    const targetDir = '/home/null/Desktop/gatewaytools/dot-bypasser-4.2.0-chrome';
    const results = {
        permissions: {},
        urls: [],
        apis: [],
        encodedStrings: [],
        suspiciousPatterns: [],
        networkActivity: [],
        remainingObfuscation: [],
        securityConcerns: []
    };

    console.log('🔍 COMPREHENSIVE SECURITY ANALYSIS');
    console.log('==================================\n');

    // 1. Analyze manifest permissions
    console.log('1. PERMISSION ANALYSIS');
    console.log('----------------------');
    try {
        const manifest = JSON.parse(fs.readFileSync(path.join(targetDir, 'manifest.json'), 'utf8'));
        results.permissions = manifest.permissions;
        results.hostPermissions = manifest.host_permissions;
        
        console.log('🚨 CRITICAL PERMISSIONS:');
        manifest.permissions.forEach(perm => {
            const risk = analyzePermission(perm);
            if (risk.level === 'HIGH') {
                console.log(`  ❌ ${perm} - ${risk.description}`);
                results.securityConcerns.push(`Permission: ${perm} - ${risk.description}`);
            }
        });
        
        if (manifest.host_permissions && manifest.host_permissions.includes('<all_urls>')) {
            console.log('  ❌ <all_urls> - ACCESS TO ALL WEBSITES');
            results.securityConcerns.push('Host Permission: <all_urls> - Full web access');
        }
    } catch (e) {
        console.log('❌ Could not analyze manifest:', e.message);
    }

    // 2. Search for URLs and API endpoints
    console.log('\n2. URL AND API ANALYSIS');
    console.log('------------------------');
    const deobfuscatedFiles = [
        'background_advanced_deobfuscated.js',
        'content-scripts/main_advanced_deobfuscated.js',
        'chunks/popup-DcHIow2Q_advanced_deobfuscated.js'
    ];

    deobfuscatedFiles.forEach(file => {
        const filePath = path.join(targetDir, file);
        if (fs.existsSync(filePath)) {
            console.log(`\n📁 Analyzing ${file}:`);
            const content = fs.readFileSync(filePath, 'utf8');
            
            // Search for URLs
            const urlPattern = /https?:\/\/[^\s"'`<>()[\]{}|\\^]+/g;
            const urls = content.match(urlPattern) || [];
            urls.forEach(url => {
                if (!results.urls.includes(url)) {
                    results.urls.push(url);
                    console.log(`  🌐 URL: ${url}`);
                }
            });

            // Search for API calls
            const apiPatterns = [
                /chrome\.runtime\.[a-zA-Z]+/g,
                /chrome\.webRequest\.[a-zA-Z]+/g,
                /chrome\.proxy\.[a-zA-Z]+/g,
                /chrome\.tabs\.[a-zA-Z]+/g,
                /fetch\s*\(/g,
                /XMLHttpRequest/g,
                /\.send\s*\(/g
            ];

            apiPatterns.forEach(pattern => {
                const matches = content.match(pattern) || [];
                matches.forEach(match => {
                    if (!results.apis.includes(match)) {
                        results.apis.push(match);
                        console.log(`  🔌 API: ${match}`);
                    }
                });
            });

            // Search for suspicious patterns
            const suspiciousPatterns = [
                { pattern: /eval\s*\(/g, name: 'eval() usage' },
                { pattern: /Function\s*\(/g, name: 'Function() constructor' },
                { pattern: /document\.write/g, name: 'document.write' },
                { pattern: /innerHTML\s*=/g, name: 'innerHTML assignment' },
                { pattern: /createElement\s*\(/g, name: 'DOM element creation' },
                { pattern: /appendChild/g, name: 'DOM manipulation' },
                { pattern: /atob\s*\(/g, name: 'Base64 decoding' },
                { pattern: /btoa\s*\(/g, name: 'Base64 encoding' },
                { pattern: /String\.fromCharCode/g, name: 'Character code conversion' }
            ];

            suspiciousPatterns.forEach(({ pattern, name }) => {
                const matches = content.match(pattern) || [];
                if (matches.length > 0) {
                    const entry = `${name} (${matches.length} occurrences)`;
                    if (!results.suspiciousPatterns.includes(entry)) {
                        results.suspiciousPatterns.push(entry);
                        console.log(`  ⚠️  ${entry}`);
                    }
                }
            });
        }
    });

    // 3. Check for remaining obfuscation
    console.log('\n3. REMAINING OBFUSCATION ANALYSIS');
    console.log('----------------------------------');
    
    deobfuscatedFiles.forEach(file => {
        const filePath = path.join(targetDir, file);
        if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf8');
            
            // Check for hex strings
            const hexStrings = content.match(/0x[0-9a-fA-F]+/g) || [];
            if (hexStrings.length > 100) {
                const concern = `${file}: ${hexStrings.length} hex values (heavily obfuscated)`;
                results.remainingObfuscation.push(concern);
                console.log(`  🔢 ${concern}`);
            }

            // Check for array lookups
            const arrayLookups = content.match(/\w+\[0x[0-9a-fA-F]+\]/g) || [];
            if (arrayLookups.length > 50) {
                const concern = `${file}: ${arrayLookups.length} array lookups (obfuscated strings)`;
                results.remainingObfuscation.push(concern);
                console.log(`  📚 ${concern}`);
            }

            // Check for encoded strings
            const encodedStrings = content.match(/\\x[0-9a-fA-F]{2}/g) || [];
            if (encodedStrings.length > 10) {
                const concern = `${file}: ${encodedStrings.length} hex-encoded characters`;
                results.remainingObfuscation.push(concern);
                console.log(`  🔤 ${concern}`);
            }
        }
    });

    // 4. Security assessment
    console.log('\n4. SECURITY ASSESSMENT');
    console.log('======================');
    
    let riskScore = 0;
    const criticalFindings = [];

    // Manifest analysis
    if (results.permissions.includes('proxy')) {
        riskScore += 25;
        criticalFindings.push('🚨 PROXY permission - Can intercept ALL network traffic');
    }
    if (results.permissions.includes('webRequest')) {
        riskScore += 20;
        criticalFindings.push('🚨 WEBREQUEST permission - Can monitor web requests');
    }
    if (results.hostPermissions && results.hostPermissions.includes('<all_urls>')) {
        riskScore += 25;
        criticalFindings.push('🚨 ALL_URLS permission - Access to every website');
    }
    if (results.permissions.includes('scripting')) {
        riskScore += 15;
        criticalFindings.push('⚠️ SCRIPTING permission - Can inject code');
    }

    // Code analysis
    if (results.suspiciousPatterns.some(p => p.includes('eval'))) {
        riskScore += 15;
        criticalFindings.push('⚠️ Uses eval() - Dynamic code execution');
    }
    if (results.suspiciousPatterns.some(p => p.includes('Function'))) {
        riskScore += 10;
        criticalFindings.push('⚠️ Uses Function() constructor - Code generation');
    }
    if (results.remainingObfuscation.length > 0) {
        riskScore += 20;
        criticalFindings.push('🚨 Heavily obfuscated code - Hides functionality');
    }

    console.log('\n📊 RISK ASSESSMENT:');
    console.log(`Overall Risk Score: ${riskScore}/100`);
    
    if (riskScore >= 80) {
        console.log('🚨 EXTREME RISK - This extension is HIGHLY DANGEROUS');
    } else if (riskScore >= 60) {
        console.log('⚠️ HIGH RISK - This extension poses significant security threats');
    } else if (riskScore >= 40) {
        console.log('⚠️ MODERATE RISK - This extension has concerning capabilities');
    } else {
        console.log('✅ LOW RISK - This extension appears relatively safe');
    }

    console.log('\n🔍 CRITICAL FINDINGS:');
    criticalFindings.forEach(finding => console.log(`  ${finding}`));

    console.log('\n📋 SUMMARY:');
    console.log(`  • ${results.permissions.length} permissions granted`);
    console.log(`  • ${results.urls.length} URLs found`);
    console.log(`  • ${results.apis.length} API calls identified`);  
    console.log(`  • ${results.suspiciousPatterns.length} suspicious patterns`);
    console.log(`  • ${results.remainingObfuscation.length} obfuscation indicators`);
    console.log(`  • ${results.securityConcerns.length} security concerns`);

    // Final recommendation
    console.log('\n🎯 RECOMMENDATION:');
    console.log('This extension appears to be MALICIOUS SOFTWARE designed for:');
    console.log('  • Payment fraud (CVV bypassing)');
    console.log('  • Network traffic interception');
    console.log('  • Unauthorized data collection');
    console.log('  • Code obfuscation to hide malicious intent');
    console.log('\n❌ DO NOT INSTALL OR USE THIS EXTENSION');
    console.log('❌ REMOVE IMMEDIATELY IF ALREADY INSTALLED');
    console.log('❌ REPORT TO SECURITY AUTHORITIES IF NEEDED');

    return results;
}

function analyzePermission(permission) {
    const permissionRisks = {
        'proxy': { level: 'HIGH', description: 'Can intercept and modify ALL network traffic' },
        'webRequest': { level: 'HIGH', description: 'Can monitor and block web requests' },
        'webRequestAuthProvider': { level: 'HIGH', description: 'Can handle authentication requests' },
        'tabs': { level: 'MEDIUM', description: 'Can access browser tab information' },
        'scripting': { level: 'HIGH', description: 'Can inject arbitrary code into web pages' },
        'storage': { level: 'LOW', description: 'Can store data locally' },
        'alarms': { level: 'LOW', description: 'Can set periodic alarms' },
        'declarativeNetRequestWithHostAccess': { level: 'HIGH', description: 'Can modify network requests' },
        'webNavigation': { level: 'MEDIUM', description: 'Can track navigation events' },
        'offscreen': { level: 'LOW', description: 'Can create offscreen documents' },
        'system.cpu': { level: 'MEDIUM', description: 'Can access CPU information' },
        'system.memory': { level: 'MEDIUM', description: 'Can access memory information' }
    };
    
    return permissionRisks[permission] || { level: 'UNKNOWN', description: 'Unknown permission' };
}

if (require.main === module) {
    comprehensiveAnalysis();
}

module.exports = { comprehensiveAnalysis };
