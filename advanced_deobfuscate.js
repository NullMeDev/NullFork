#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// LZ-String implementation (basic decompression)
class LZString {
    static decompressFromUTF16(compressed) {
        if (compressed == null) return "";
        if (compressed == "") return null;
        return this._decompress(compressed.length, 16384, function(index) {
            return compressed.charCodeAt(index) - 32;
        });
    }

    static _decompress(length, resetValue, getNextValue) {
        let dictionary = {},
            next,
            enlargeIn = 4,
            dictSize = 4,
            numBits = 3,
            entry = "",
            result = [],
            i,
            w,
            bits, resb, maxpower, power,
            c,
            data = {val: getNextValue(0), position: resetValue, index: 1};

        for (i = 0; i < 3; i += 1) {
            dictionary[i] = i;
        }

        bits = 0;
        maxpower = Math.pow(2, 2);
        power = 1;
        while (power != maxpower) {
            resb = data.val & data.position;
            data.position >>= 1;
            if (data.position == 0) {
                data.position = resetValue;
                data.val = getNextValue(data.index++);
            }
            bits |= (resb > 0 ? 1 : 0) * power;
            power <<= 1;
        }

        switch (next = bits) {
            case 0:
                bits = 0;
                maxpower = Math.pow(2, 8);
                power = 1;
                while (power != maxpower) {
                    resb = data.val & data.position;
                    data.position >>= 1;
                    if (data.position == 0) {
                        data.position = resetValue;
                        data.val = getNextValue(data.index++);
                    }
                    bits |= (resb > 0 ? 1 : 0) * power;
                    power <<= 1;
                }
                c = String.fromCharCode(bits);
                break;
            case 1:
                bits = 0;
                maxpower = Math.pow(2, 16);
                power = 1;
                while (power != maxpower) {
                    resb = data.val & data.position;
                    data.position >>= 1;
                    if (data.position == 0) {
                        data.position = resetValue;
                        data.val = getNextValue(data.index++);
                    }
                    bits |= (resb > 0 ? 1 : 0) * power;
                    power <<= 1;
                }
                c = String.fromCharCode(bits);
                break;
            case 2:
                return "";
        }
        dictionary[3] = c;
        w = c;
        result.push(c);
        while (true) {
            if (data.index > length) {
                return "";
            }

            bits = 0;
            maxpower = Math.pow(2, numBits);
            power = 1;
            while (power != maxpower) {
                resb = data.val & data.position;
                data.position >>= 1;
                if (data.position == 0) {
                    data.position = resetValue;
                    data.val = getNextValue(data.index++);
                }
                bits |= (resb > 0 ? 1 : 0) * power;
                power <<= 1;
            }

            switch (c = bits) {
                case 0:
                    bits = 0;
                    maxpower = Math.pow(2, 8);
                    power = 1;
                    while (power != maxpower) {
                        resb = data.val & data.position;
                        data.position >>= 1;
                        if (data.position == 0) {
                            data.position = resetValue;
                            data.val = getNextValue(data.index++);
                        }
                        bits |= (resb > 0 ? 1 : 0) * power;
                        power <<= 1;
                    }

                    dictionary[dictSize++] = String.fromCharCode(bits);
                    c = dictSize - 1;
                    enlargeIn--;
                    break;
                case 1:
                    bits = 0;
                    maxpower = Math.pow(2, 16);
                    power = 1;
                    while (power != maxpower) {
                        resb = data.val & data.position;
                        data.position >>= 1;
                        if (data.position == 0) {
                            data.position = resetValue;
                            data.val = getNextValue(data.index++);
                        }
                        bits |= (resb > 0 ? 1 : 0) * power;
                        power <<= 1;
                    }
                    dictionary[dictSize++] = String.fromCharCode(bits);
                    c = dictSize - 1;
                    enlargeIn--;
                    break;
                case 2:
                    return result.join('');
            }

            if (enlargeIn == 0) {
                enlargeIn = Math.pow(2, numBits);
                numBits++;
            }

            if (dictionary[c]) {
                entry = dictionary[c];
            } else {
                if (c === dictSize) {
                    entry = w + w.charAt(0);
                } else {
                    return null;
                }
            }
            result.push(entry);

            dictionary[dictSize++] = w + entry.charAt(0);
            enlargeIn--;

            w = entry;

            if (enlargeIn == 0) {
                enlargeIn = Math.pow(2, numBits);
                numBits++;
            }
        }
    }
}

// Advanced string deobfuscation
function advancedStringDecode(code) {
    // Handle different string encoding patterns
    let decoded = code;
    
    // Decode hex strings
    decoded = decoded.replace(/\\x([0-9A-Fa-f]{2})/g, (match, p1) => {
        return String.fromCharCode(parseInt(p1, 16));
    });
    
    // Decode unicode strings
    decoded = decoded.replace(/\\u([0-9A-Fa-f]{4})/g, (match, p1) => {
        return String.fromCharCode(parseInt(p1, 16));
    });
    
    // Handle string concatenation patterns
    decoded = decoded.replace(/(".*?")\s*\+\s*(".*?")/g, (match, str1, str2) => {
        try {
            return JSON.stringify(JSON.parse(str1) + JSON.parse(str2));
        } catch (e) {
            return match;
        }
    });
    
    return decoded;
}

// Extract and resolve string arrays
function extractStringArrays(code) {
    const arrays = {};
    const arrayPattern = /(?:const|var|let)\s+(\w+)\s*=\s*\[((?:"[^"]*"|'[^']*'|`[^`]*`,?\s*)*)\]/g;
    
    let match;
    while ((match = arrayPattern.exec(code)) !== null) {
        const arrayName = match[1];
        const arrayContent = match[2];
        
        // Parse array elements
        const elements = [];
        const elementPattern = /["'`]([^"'`]*)["'`]/g;
        let elementMatch;
        while ((elementMatch = elementPattern.exec(arrayContent)) !== null) {
            elements.push(elementMatch[1]);
        }
        
        if (elements.length > 5) { // Only consider substantial arrays
            arrays[arrayName] = elements;
            console.log(`Found string array ${arrayName} with ${elements.length} elements`);
        }
    }
    
    return arrays;
}

// Replace array lookups with actual strings
function replaceArrayLookups(code, arrays) {
    let result = code;
    
    for (const [arrayName, elements] of Object.entries(arrays)) {
        // Replace hex index lookups
        const hexPattern = new RegExp(`${arrayName}\\[0x([0-9A-Fa-f]+)\\]`, 'g');
        result = result.replace(hexPattern, (match, hexIndex) => {
            const index = parseInt(hexIndex, 16);
            if (index < elements.length) {
                return `"${elements[index]}" /* was ${match} */`;
            }
            return match;
        });
        
        // Replace decimal index lookups
        const decPattern = new RegExp(`${arrayName}\\[([0-9]+)\\]`, 'g');
        result = result.replace(decPattern, (match, decIndex) => {
            const index = parseInt(decIndex, 10);
            if (index < elements.length) {
                return `"${elements[index]}" /* was ${match} */`;
            }
            return match;
        });
    }
    
    return result;
}

// Attempt to detect and decode LZ-String compressed data
function detectAndDecodeLZString(code) {
    // Look for patterns that might be LZ-String compressed data
    const lzPatterns = [
        /decompressFromUTF16\s*\(\s*["']([^"']+)["']\s*\)/g,
        /["']([^\s"']{100,})["']/g // Long strings without spaces might be compressed
    ];
    
    let result = code;
    
    for (const pattern of lzPatterns) {
        result = result.replace(pattern, (match, compressed) => {
            try {
                const decompressed = LZString.decompressFromUTF16(compressed);
                if (decompressed && decompressed.length > compressed.length / 2) {
                    console.log(`Decompressed LZ-String: ${compressed.substring(0, 50)}... -> ${decompressed.substring(0, 100)}...`);
                    return `"${decompressed.replace(/"/g, '\\"')}" /* decompressed from LZ-String */`;
                }
            } catch (e) {
                // Not LZ-String compressed, leave as is
            }
            return match;
        });
    }
    
    return result;
}

// Function name deobfuscation
function deobfuscateFunctionNames(code) {
    // Try to identify common patterns and rename functions
    const patterns = [
        {
            pattern: /function\s+[a-zA-Z_$][\w$]*\s*\([^)]*\)\s*{\s*for\s*\([^)]*\)\s*[\w\[\]\.]+\.push\s*\([\w\[\]\.]+\.shift\s*\(\s*\)\s*\)/g,
            name: 'arrayRotateFunction'
        },
        {
            pattern: /function\s+[a-zA-Z_$][\w$]*\s*\([^)]*\)\s*{\s*return\s+String\.fromCharCode/g,
            name: 'stringDecodeFunction'
        },
        {
            pattern: /function\s+[a-zA-Z_$][\w$]*\s*\([^)]*\)\s*{\s*.*String\.prototype\.charAt/g,
            name: 'stringCharFunction'
        }
    ];
    
    let result = code;
    let counter = 1;
    
    for (const {pattern, name} of patterns) {
        result = result.replace(pattern, (match) => {
            return match.replace(/function\s+([a-zA-Z_$][\w$]*)/, `function ${name}_${counter++} /* was $1 */`);
        });
    }
    
    return result;
}

// Enhanced beautification
function enhancedBeautify(code) {
    let beautified = code;
    
    // Add line breaks and basic formatting
    beautified = beautified
        .replace(/;(?!\s*[\n\r])/g, ';\n')
        .replace(/\{(?!\s*[\n\r])/g, '{\n')
        .replace(/\}(?!\s*[\n\r])/g, '\n}\n')
        .replace(/,(?!\s*[\n\r])/g, ',\n')
        .replace(/&&/g, ' && ')
        .replace(/\|\|/g, ' || ')
        .replace(/==/g, ' == ')
        .replace(/!=/g, ' != ')
        .replace(/===/g, ' === ')
        .replace(/!==/g, ' !== ');
    
    // Basic indentation
    const lines = beautified.split('\n');
    let indent = 0;
    const indented = lines.map(line => {
        const trimmed = line.trim();
        if (trimmed === '') return '';
        
        // Decrease indent for closing braces
        if (trimmed.includes('}') && !trimmed.includes('{')) {
            indent = Math.max(0, indent - 1);
        }
        
        const result = '  '.repeat(indent) + trimmed;
        
        // Increase indent for opening braces
        if (trimmed.includes('{') && !trimmed.includes('}')) {
            indent++;
        }
        
        return result;
    });
    
    return indented.join('\n');
}

// Main advanced deobfuscation function
function advancedDeobfuscate(filePath) {
    console.log(`\n=== ADVANCED DEOBFUSCATING: ${path.basename(filePath)} ===`);
    
    try {
        let code = fs.readFileSync(filePath, 'utf8');
        const originalSize = code.length;
        
        console.log(`Original size: ${originalSize} characters`);
        
        // Step 1: Advanced string decoding
        console.log('Step 1: Advanced string decoding...');
        code = advancedStringDecode(code);
        
        // Step 2: Detect and decode LZ-String
        console.log('Step 2: Detecting and decoding LZ-String...');
        code = detectAndDecodeLZString(code);
        
        // Step 3: Extract string arrays
        console.log('Step 3: Extracting string arrays...');
        const stringArrays = extractStringArrays(code);
        
        // Step 4: Replace array lookups
        console.log('Step 4: Replacing array lookups...');
        code = replaceArrayLookups(code, stringArrays);
        
        // Step 5: Deobfuscate function names
        console.log('Step 5: Deobfuscating function names...');
        code = deobfuscateFunctionNames(code);
        
        // Step 6: Enhanced beautification
        console.log('Step 6: Enhanced beautification...');
        code = enhancedBeautify(code);
        
        // Save the result
        const outputPath = filePath.replace(/\.js$/, '_advanced_deobfuscated.js');
        fs.writeFileSync(outputPath, code);
        
        console.log(`Advanced deobfuscated file saved to: ${outputPath}`);
        console.log(`New size: ${code.length} characters`);
        console.log(`Size ratio: ${(code.length / originalSize * 100).toFixed(1)}%`);
        
        // Show sample of the result
        console.log('\nFirst 800 characters of advanced deobfuscated code:');
        console.log('-'.repeat(60));
        console.log(code.substring(0, 800));
        console.log('-'.repeat(60));
        
        return outputPath;
        
    } catch (error) {
        console.error(`Error in advanced deobfuscation of ${filePath}:`, error.message);
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
    
    console.log('Advanced JavaScript Deobfuscator');
    console.log('==================================');
    
    // Focus on the most important files first
    const priorityFiles = [
        'content-scripts/main.js',
        'background.js',
        'chunks/popup-DcHIow2Q.js'
    ];
    
    const results = [];
    
    for (const file of priorityFiles) {
        const fullPath = path.join(targetDir, file);
        if (fs.existsSync(fullPath)) {
            const result = advancedDeobfuscate(fullPath);
            if (result) {
                results.push(result);
            }
        } else {
            console.log(`File not found: ${fullPath}`);
        }
    }
    
    console.log(`\n=== ADVANCED DEOBFUSCATION COMPLETE ===`);
    console.log(`Successfully processed ${results.length}/${priorityFiles.length} priority files`);
    
    if (results.length > 0) {
        console.log('\nAdvanced deobfuscated files:');
        results.forEach(file => console.log(`  - ${file}`));
        
        console.log('\nRecommendations:');
        console.log('1. Review the deobfuscated files for any remaining obfuscation');
        console.log('2. Look for API calls, URLs, and suspicious functionality');
        console.log('3. Check for any remaining compressed strings or encoded data');
        console.log('4. Analyze the extension\'s behavior and permissions');
    }
}

if (require.main === module) {
    main();
}

module.exports = { advancedDeobfuscate, LZString };
