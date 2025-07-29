#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Simple LZ-String decompression (basic implementation)
function decompressLZString(compressed) {
    try {
        // This is a basic attempt - real LZ-String is more complex
        if (compressed.includes('\\u')) {
            return compressed.replace(/\\u([0-9A-Fa-f]{4})/g, (match, p1) => {
                return String.fromCharCode(parseInt(p1, 16));
            });
        }
    } catch (e) {
        console.warn('LZ-String decompression failed:', e.message);
    }
    return compressed;
}

// Decode hex strings
function decodeHexStrings(code) {
    // Replace hex encoded strings like \x41\x42\x43
    return code.replace(/\\x([0-9A-Fa-f]{2})/g, (match, p1) => {
        return String.fromCharCode(parseInt(p1, 16));
    });
}

// Decode unicode strings
function decodeUnicodeStrings(code) {
    // Replace unicode encoded strings like \u0041\u0042\u0043
    return code.replace(/\\u([0-9A-Fa-f]{4})/g, (match, p1) => {
        return String.fromCharCode(parseInt(p1, 16));
    });
}

// Try to resolve array-based string lookups
function resolveArrayLookups(code) {
    // Look for patterns like arr[0x1a] or arr[26]
    const arrayPattern = /(\w+)\[0x([0-9A-Fa-f]+)\]/g;
    return code.replace(arrayPattern, (match, arrayName, hexIndex) => {
        const index = parseInt(hexIndex, 16);
        return `${arrayName}[${index}] /* was ${match} */`;
    });
}

// Beautify basic JavaScript structure
function beautifyCode(code) {
    // Add line breaks after semicolons and braces
    let beautified = code
        .replace(/;/g, ';\n')
        .replace(/\{/g, '{\n')
        .replace(/\}/g, '\n}')
        .replace(/,/g, ',\n');
    
    // Basic indentation
    const lines = beautified.split('\n');
    let indent = 0;
    const indented = lines.map(line => {
        const trimmed = line.trim();
        if (trimmed.includes('}')) indent = Math.max(0, indent - 1);
        const result = '  '.repeat(indent) + trimmed;
        if (trimmed.includes('{')) indent++;
        return result;
    });
    
    return indented.join('\n');
}

// Try to detect and extract function parameters/content
function analyzeObfuscatedFunction(code) {
    const analysis = {
        stringArrays: [],
        functions: [],
        variables: [],
        decodingFunctions: []
    };
    
    // Look for large string arrays (potential lookup tables)
    const stringArrayPattern = /(\w+)\s*=\s*\[(["'][^"']*["'],?\s*){10,}\]/g;
    let match;
    while ((match = stringArrayPattern.exec(code)) !== null) {
        analysis.stringArrays.push({
            name: match[1],
            position: match.index
        });
    }
    
    // Look for function definitions
    const functionPattern = /function\s+(\w+)\s*\([^)]*\)\s*\{/g;
    while ((match = functionPattern.exec(code)) !== null) {
        analysis.functions.push({
            name: match[1],
            position: match.index
        });
    }
    
    return analysis;
}

// Main deobfuscation function
function deobfuscateFile(filePath) {
    console.log(`\nDeobfuscating: ${filePath}`);
    
    try {
        let code = fs.readFileSync(filePath, 'utf8');
        const originalSize = code.length;
        
        console.log(`Original size: ${originalSize} characters`);
        
        // Step 1: Decode hex strings
        console.log('Step 1: Decoding hex strings...');
        code = decodeHexStrings(code);
        
        // Step 2: Decode unicode strings  
        console.log('Step 2: Decoding unicode strings...');
        code = decodeUnicodeStrings(code);
        
        // Step 3: Attempt LZ-String decompression
        console.log('Step 3: Attempting LZ-String decompression...');
        code = decompressLZString(code);
        
        // Step 4: Resolve array lookups
        console.log('Step 4: Resolving array lookups...');
        code = resolveArrayLookups(code);
        
        // Step 5: Analyze structure
        console.log('Step 5: Analyzing obfuscated structure...');
        const analysis = analyzeObfuscatedFunction(code);
        console.log(`Found ${analysis.stringArrays.length} string arrays`);
        console.log(`Found ${analysis.functions.length} functions`);
        
        // Step 6: Beautify
        console.log('Step 6: Beautifying code...');
        code = beautifyCode(code);
        
        // Save deobfuscated file
        const outputPath = filePath.replace(/\.js$/, '_deobfuscated.js');
        fs.writeFileSync(outputPath, code);
        
        console.log(`Deobfuscated file saved to: ${outputPath}`);
        console.log(`New size: ${code.length} characters`);
        
        // Show first 500 characters of result
        console.log('\nFirst 500 characters of deobfuscated code:');
        console.log('-'.repeat(50));
        console.log(code.substring(0, 500));
        console.log('-'.repeat(50));
        
        return outputPath;
        
    } catch (error) {
        console.error(`Error deobfuscating ${filePath}:`, error.message);
        return null;
    }
}

// Main execution
function main() {
    const targetDir = '/home/null/Desktop/gatewaytools/dot-bypasser-4.2.0-chrome';
    
    if (!fs.existsSync(targetDir)) {
        console.error(`Target directory not found: ${targetDir}`);
        process.exit(1);
    }
    
    console.log('JavaScript Deobfuscator');
    console.log('========================');
    console.log(`Target directory: ${targetDir}`);
    
    // Find all JavaScript files
    function findJSFiles(dir) {
        const files = [];
        const items = fs.readdirSync(dir);
        
        for (const item of items) {
            const fullPath = path.join(dir, item);
            const stat = fs.statSync(fullPath);
            
            if (stat.isDirectory()) {
                files.push(...findJSFiles(fullPath));
            } else if (item.endsWith('.js') && !item.includes('_deobfuscated')) {
                files.push(fullPath);
            }
        }
        
        return files;
    }
    
    const jsFiles = findJSFiles(targetDir);
    console.log(`\nFound ${jsFiles.length} JavaScript files to process:`);
    jsFiles.forEach(file => console.log(`  - ${path.relative(targetDir, file)}`));
    
    // Process each file
    const results = [];
    for (const file of jsFiles) {
        const result = deobfuscateFile(file);
        if (result) {
            results.push(result);
        }
    }
    
    console.log(`\n=== DEOBFUSCATION COMPLETE ===`);
    console.log(`Successfully processed ${results.length}/${jsFiles.length} files`);
    console.log('\nDeobfuscated files:');
    results.forEach(file => console.log(`  - ${file}`));
}

if (require.main === module) {
    main();
}

module.exports = {
    deobfuscateFile,
    decodeHexStrings,
    decodeUnicodeStrings,
    beautifyCode
};
