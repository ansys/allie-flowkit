import os
import shutil
import re

# Set environment variables for file paths
DOC_BUILD_HTML = "documentation-html"
SOURCE_FILE = os.path.join(DOC_BUILD_HTML, "api_reference/test/index.html")
SOURCE_DIRECTORY = "dist/pkg/github.com/ansys/allie-flowkit/pkg/"
REPLACEMENT_DIRECTORY = os.path.join(DOC_BUILD_HTML, "api_reference/pkg")
ACTUAL_DIR = os.path.join(DOC_BUILD_HTML, "api_reference")

# Check if REPLACEMENT_DIRECTORY exists, if not, create it
if not os.path.exists(REPLACEMENT_DIRECTORY):
    os.makedirs(REPLACEMENT_DIRECTORY)

# Remove existing content in the replacement directory
for filename in os.listdir(REPLACEMENT_DIRECTORY):
    file_path = os.path.join(REPLACEMENT_DIRECTORY, filename)
    if os.path.isfile(file_path) or os.path.islink(file_path):
        os.unlink(file_path)
    elif os.path.isdir(file_path):
        shutil.rmtree(file_path)

# Move the source_directory content to the replacement_directory
for filename in os.listdir(SOURCE_DIRECTORY):
    source_path = os.path.join(SOURCE_DIRECTORY, filename)
    destination_path = os.path.join(REPLACEMENT_DIRECTORY, filename)
    shutil.move(source_path, destination_path)

# Remove the index.html file in the replacement_directory
index_html_path = os.path.join(REPLACEMENT_DIRECTORY, "index.html")
if os.path.exists(index_html_path):
    os.remove(index_html_path)

# Process each HTML file in the replacement directory
for root, _, files in os.walk(REPLACEMENT_DIRECTORY):
    for file in files:
        if file.endswith(".html"):
            replacement_file = os.path.join(root, file)
            with open(replacement_file, 'r') as f:
                content = f.read()

            # Extract the body content
            body_content_match = re.search(r'<body>(.*?)</body>', content, re.DOTALL)
            if body_content_match:
                body_content = body_content_match.group(1)

                # Remove specific div and its content
                body_content = re.sub(r'<div class="top-heading" id="heading-wide"><a href="\/pkg\/github.com\/ansys\/allie-flowkit\/">GoPages \| Auto-generated docs<\/a><\/div>.*?<a href="#" id="menu-button"><span id="menu-button-arrow">&#9661;<\/span><\/a>', '', body_content, flags=re.DOTALL)

                # Escape slashes
                body_content = body_content.replace('/', r'\/')

                # Replace content between specific HTML tags using a custom logic
                def replace_content(match):
                    return match.group(1) + body_content + match.group(3)

                content = re.sub(r'(<article class="bd-article" role="main">).*?(</article>)', replace_content, content, flags=re.DOTALL)

                with open(replacement_file, 'w') as f:
                    f.write(content)

# Move the modified files back to the actual directory
for filename in os.listdir(REPLACEMENT_DIRECTORY):
    source_path = os.path.join(REPLACEMENT_DIRECTORY, filename)
    destination_path = os.path.join(ACTUAL_DIR, filename)
    shutil.move(source_path, destination_path)
