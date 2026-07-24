#!/usr/bin/env python3
import sys
import zipfile
import xml.etree.ElementTree as ET

def validate_docx(filepath):
    """
    Performs a structural validation of a DOCX file by verifying its ZIP integrity
    and the well-formedness of its core XML components.
    """
    try:
        # 1. Archive Integrity Check
        # This verifies the central directory structure of the ZIP is intact.
        with zipfile.ZipFile(filepath, 'r') as zf:
            
            # 2. OOXML Schema Presence Check
            # A valid DOCX must contain these minimum components.
            required_files = ['word/document.xml', '[Content_Types].xml', 'word/_rels/document.xml.rels']
            archive_files = zf.namelist()
            
            for req_file in required_files:
                if req_file not in archive_files:
                    print(f"ERROR: Structural failure. Missing required OOXML component: '{req_file}'")
                    return False

            # 3. XML Well-Formedness Check
            # We specifically target the document payload mutated by the Go redactor.
            xml_content = zf.read('word/document.xml')
            
            try:
                # Attempt to parse the byte stream into an XML Abstract Syntax Tree.
                # This will catch missing closing tags, unescaped reserved characters, and structural corruption.
                ET.fromstring(xml_content)
            except ET.ParseError as e:
                print(f"ERROR: Malformed XML detected in 'word/document.xml'.")
                print(f"Parser Output: {e}")
                return False

    except zipfile.BadZipFile:
        print(f"ERROR: {filepath} is not a valid ZIP archive. The repackaging stage likely failed.")
        return False
    except Exception as e:
        print(f"ERROR: An unexpected systemic failure occurred during validation: {e}")
        return False

    print(f"SUCCESS: {filepath} passed strict structural validation.")
    return True

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 validate.py <path_to_docx>")
        sys.exit(1)

    # Standard exit codes for pipeline orchestration: 0 for success, 1 for failure.
    target_file = sys.argv[1]
    if validate_docx(target_file):
        sys.exit(0)
    else:
        sys.exit(1)