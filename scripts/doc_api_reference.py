import os
import shutil
from pathlib import Path
import sys
from bs4 import BeautifulSoup, Tag

# Set environment variables for file paths
DOC_BUILD_HTML = "documentation-html"
SOURCE_FILE = f"{DOC_BUILD_HTML}/api_reference/test/index.html"
SOURCE_DIRECTORY = "dist/pkg/github.com/ansys/allie-flowkit/pkg/"
REPLACEMENT_DIRECTORY = f"{DOC_BUILD_HTML}/api_reference/pkg"
ACTUAL_DIR = f"{DOC_BUILD_HTML}/api_reference/"

# Check if REPLACEMENT_DIRECTORY exists, if not, create it
os.makedirs(REPLACEMENT_DIRECTORY, exist_ok=True)

# Remove existing content in the replacement directory
shutil.rmtree(REPLACEMENT_DIRECTORY)
os.makedirs(REPLACEMENT_DIRECTORY)

# Move the source_directory content to the replacement_directory
for item in Path(SOURCE_DIRECTORY).glob('*'):
    shutil.move(str(item), REPLACEMENT_DIRECTORY)

# Remove the index.html file in the replacement_directory
index_file_path = Path(REPLACEMENT_DIRECTORY) / "index.html"
if index_file_path.exists():
    index_file_path.unlink()

# Process each HTML file in the replacement directory
for replacement_file_path in Path(REPLACEMENT_DIRECTORY).glob("*.html"):
    with open(replacement_file_path, 'r', encoding='utf-8') as file:
        soup = BeautifulSoup(file, 'lxml')

    # Extract body content and clean up specific HTML tags
    if soup.body:
        body_content = soup.body.extract()
        top_heading = body_content.find("div", {"class": "top-heading"})
        if top_heading and isinstance(top_heading, Tag):
            top_heading.decompose()

        menu_button = body_content.find("a", {"id": "menu-button"})
        if menu_button and isinstance(menu_button, Tag):
            menu_button.decompose()

        replacement_body_content = str(body_content).replace('/', r'\/')

    # Read the source file and replace the content between specific HTML tags
    with open(SOURCE_FILE, 'r', encoding='utf-8') as source_file:
        source_soup = BeautifulSoup(source_file, 'lxml')
        article_tag = source_soup.find("article", {"class": "bd-article", "role": "main"})
        if article_tag and isinstance(article_tag, Tag):
            article_tag.clear()
            article_tag.append(BeautifulSoup(replacement_body_content, 'lxml'))

    # Write the modified content back to the replacement file
    with open(replacement_file_path, 'w', encoding='utf-8') as file:
        file.write(str(source_soup))

# Move the modified files back to the actual directory
for item in Path(REPLACEMENT_DIRECTORY).glob('*'):
    shutil.move(str(item), ACTUAL_DIR)