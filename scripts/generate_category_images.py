#!/usr/bin/env python3
"""
Script to generate 300x300 pixel category images with 1:1 aspect ratio
"""
from PIL import Image, ImageDraw, ImageFont
import os
import sys

# Category data
categories = [
    (1, "Fasteners", 1), (2, "Plumbing", 1), (3, "Electrical", 1), (4, "Paints", 1),
    (5, "Power Tools", 1), (6, "Hand Tools", 1), (7, "Hardware Accessories", 1),
    (8, "Ladders & Platforms", 1), (9, "Measuring Tools", 1), (10, "Cutting Tools", 1),
    (11, "Safety Equipment", 1), (12, "Garden Tools", 1), (13, "Welding Supplies", 1),
    (14, "Adhesives & Sealants", 1), (15, "Others", 1), (16, "Building Materials", 2),
    (17, "Steel & Metals", 2), (18, "Wood & Timber", 2), (19, "Roofing", 2),
    (20, "Civil Supplies", 2), (21, "Cement & Concrete", 2), (22, "Sand & Aggregates", 2),
    (23, "Bricks & Blocks", 2), (24, "Doors & Windows", 2), (25, "Plywood & Laminates", 2),
    (26, "Gypsum & Plaster", 2), (27, "Tiles & Flooring", 2), (28, "Structural Steel", 2),
    (29, "Pipes & Fittings", 2), (30, "Construction Chemicals", 2), (31, "Others", 2),
    (32, "Living Room", 3), (33, "Bedroom", 3), (34, "Kitchen", 3), (35, "Office", 3),
    (36, "Outdoor", 3), (37, "Dining Room", 3), (38, "Mattresses", 3),
    (39, "Sofas & Recliners", 3), (40, "Beds", 3), (41, "Wardrobes", 3),
    (42, "Tables", 3), (43, "Chairs", 3), (44, "Cabinets & Storage", 3),
    (45, "TV Units & Entertainment", 3), (46, "Bookcases & Shelves", 3), (47, "Others", 3),
    (48, "Mobiles", 4), (49, "Tablets", 4), (50, "Laptops", 4), (51, "Desktops", 4),
    (52, "Televisions", 4), (53, "Home Appliances", 4), (54, "Kitchen Appliances", 4),
    (55, "Audio Systems", 4), (56, "Cameras", 4), (57, "Gaming Consoles", 4),
    (58, "Accessories", 4), (59, "Networking", 4), (60, "Smart Home", 4),
    (61, "Wearables", 4), (62, "Printers & Scanners", 4), (63, "Storage Devices", 4),
    (64, "Speakers", 4), (65, "Headphones", 4), (66, "Power Banks", 4),
    (67, "Fans & Coolers", 4), (68, "Others", 4), (69, "Staples", 5), (70, "Snacks", 5),
    (71, "Beverages", 5), (72, "Household", 5), (73, "Dairy", 5), (74, "Bakery", 5),
    (75, "Spices & Masalas", 5), (76, "Cooking Essentials", 5), (77, "Personal Care", 5),
    (78, "Baby Products", 5), (79, "Health & Wellness", 5), (80, "Frozen Foods", 5),
    (81, "Ready to Eat", 5), (82, "Organic Products", 5), (83, "Fruits & Vegetables", 5),
    (84, "Sweets & Mithai", 5), (85, "Pet Supplies", 5), (86, "Others", 5),
    (87, "Office Supplies", 6), (88, "School Supplies", 6), (89, "Art Materials", 6),
    (90, "Packaging", 6), (91, "Writing Instruments", 6), (92, "Notebooks & Diaries", 6),
    (93, "Files & Folders", 6), (94, "Desk Accessories", 6), (95, "Printer Supplies", 6),
    (96, "Drawing & Drafting", 6), (97, "Craft Supplies", 6), (98, "Whiteboards & Boards", 6),
    (99, "Calculators", 6), (100, "Presentation Supplies", 6), (101, "Correction Supplies", 6),
    (102, "Sticky Notes & Tapes", 6), (103, "Envelopes & Labels", 6), (104, "Others", 6),
    (105, "Shirts", 7), (106, "Pants", 7), (107, "Suits", 7), (108, "Accessory", 7),
    (109, "Dresses", 7), (110, "Tops", 7), (111, "Bottoms", 7), (112, "Lingerie", 7),
    (113, "Infants", 7), (114, "Toddlers", 7), (115, "Juniors", 7), (116, "Gym", 7),
    (117, "Outerwear", 7), (118, "Women Sportswear", 7), (119, "Mens Sportswear", 7),
    (120, "Ethnic Wear", 7), (121, "Formal Wear", 7), (122, "Casual Wear", 7),
    (123, "Sleepwear", 7), (124, "Swimwear", 7), (125, "Winter Wear", 7),
    (126, "Footwear", 7), (127, "Bags", 7), (128, "Jewelry", 7), (129, "Sunglasses", 7),
    (130, "Watches", 7), (131, "Innerwear", 7), (132, "Plus Size", 7),
    (133, "Bridal Wear", 7), (134, "Party Wear", 7), (135, "Others", 7)
]

# Color schemes for different departments
department_colors = {
    1: ("#FF6B6B", "#4ECDC4"),  # Hardware - Red/Teal
    2: ("#FFD93D", "#6BCB77"),  # Construction - Yellow/Green
    3: ("#A8DADC", "#457B9D"),  # Furniture - Light Blue/Dark Blue
    4: ("#B084CC", "#EE6C4D"),  # Electronics - Purple/Orange
    5: ("#90E0EF", "#0077B6"),  # Grocery - Light Blue/Dark Blue
    6: ("#FFB703", "#FB8500"),  # Stationery - Yellow/Orange
    7: ("#FF006E", "#8338EC"),  # Apparel - Pink/Purple
}

def generate_category_image(category_id, category_name, department_id, output_dir):
    """Generate a 300x300 image for a category"""
    # Create image with 1:1 aspect ratio (300x300)
    width, height = 300, 300
    
    # Get colors based on department
    bg_color, text_color = department_colors.get(department_id, ("#6C757D", "#F8F9FA"))
    
    # Create image with gradient background
    image = Image.new('RGB', (width, height), bg_color)
    draw = ImageDraw.Draw(image)
    
    # Add gradient effect
    for i in range(height):
        alpha = i / height
        r = int(int(bg_color[1:3], 16) * (1 - alpha) + int(text_color[1:3], 16) * alpha)
        g = int(int(bg_color[3:5], 16) * (1 - alpha) + int(text_color[3:5], 16) * alpha)
        b = int(int(bg_color[5:7], 16) * (1 - alpha) + int(text_color[5:7], 16) * alpha)
        draw.line([(0, i), (width, i)], fill=(r, g, b))
    
    # Try to use a nice font, fallback to default if not available
    try:
        font_large = ImageFont.truetype("/System/Library/Fonts/Helvetica.ttc", 36)
        font_small = ImageFont.truetype("/System/Library/Fonts/Helvetica.ttc", 18)
    except:
        try:
            font_large = ImageFont.truetype("/Library/Fonts/Arial.ttf", 36)
            font_small = ImageFont.truetype("/Library/Fonts/Arial.ttf", 18)
        except:
            font_large = ImageFont.load_default()
            font_small = ImageFont.load_default()
    
    # Split text into multiple lines if too long
    words = category_name.split()
    lines = []
    current_line = []
    
    for word in words:
        test_line = ' '.join(current_line + [word])
        bbox = draw.textbbox((0, 0), test_line, font=font_large)
        text_width = bbox[2] - bbox[0]
        
        if text_width <= width - 40:  # 20px padding on each side
            current_line.append(word)
        else:
            if current_line:
                lines.append(' '.join(current_line))
            current_line = [word]
    
    if current_line:
        lines.append(' '.join(current_line))
    
    # Calculate total text height
    line_height = 45
    total_height = len(lines) * line_height
    
    # Draw text centered
    y_position = (height - total_height) // 2
    
    for line in lines:
        bbox = draw.textbbox((0, 0), line, font=font_large)
        text_width = bbox[2] - bbox[0]
        x_position = (width - text_width) // 2
        
        # Draw text shadow for better visibility
        draw.text((x_position + 2, y_position + 2), line, fill='#000000', font=font_large)
        draw.text((x_position, y_position), line, fill='#FFFFFF', font=font_large)
        y_position += line_height
    
    # Draw category ID at bottom
    id_text = f"#{category_id}"
    bbox = draw.textbbox((0, 0), id_text, font=font_small)
    id_width = bbox[2] - bbox[0]
    draw.text(((width - id_width) // 2, height - 30), id_text, fill='#FFFFFF', font=font_small)
    
    # Save image
    filename = f"category_{category_id}.png"
    filepath = os.path.join(output_dir, filename)
    image.save(filepath, 'PNG')
    
    return filename

def main():
    # Set output directory
    output_dir = "uploads/category-images"
    
    # Create directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)
    
    print(f"Generating 300x300 images for {len(categories)} categories...")
    print(f"Output directory: {output_dir}")
    print("-" * 60)
    
    generated_count = 0
    for category_id, category_name, department_id in categories:
        try:
            filename = generate_category_image(category_id, category_name, department_id, output_dir)
            print(f"✓ Generated: {filename} - {category_name}")
            generated_count += 1
        except Exception as e:
            print(f"✗ Error generating image for {category_name}: {e}")
    
    print("-" * 60)
    print(f"Successfully generated {generated_count} out of {len(categories)} images")
    
    # Generate SQL insert statements
    print("\nGenerating SQL insert statements...")
    sql_file = "scripts/insert_category_images.sql"
    os.makedirs("scripts", exist_ok=True)
    
    with open(sql_file, 'w') as f:
        f.write("-- Insert category images\n")
        f.write("-- Generated by generate_category_images.py\n\n")
        
        for category_id, category_name, _ in categories:
            image_url = f"/uploads/category-images/category_{category_id}.png"
            alt_text = f"{category_name} Category"
            f.write(f"INSERT INTO category_images (category_id, image_url, alt_text, sort_order) \n")
            f.write(f"VALUES ({category_id}, '{image_url}', '{alt_text}', 0)\n")
            f.write(f"ON CONFLICT (category_id, image_url) DO NOTHING;\n\n")
    
    print(f"SQL file created: {sql_file}")
    print("\nDone! Run the SQL file to insert image URLs into the database.")

if __name__ == "__main__":
    main()
