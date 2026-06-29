import os

try:
    from PIL import Image, ImageDraw
    USE_PIL = True
except ImportError:
    USE_PIL = False

if not USE_PIL:
    import struct
    import zlib
    
    def png_chunk(chunk_type, data):
        chunk_len = struct.pack(">I", len(data))
        chunk_crc = struct.pack(">I", zlib.crc32(chunk_type + data) & 0xffffffff)
        return chunk_len + chunk_type + data + chunk_crc
    
    def create_png(width, height, pixels):
        signature = b'\x89PNG\r\n\x1a\n'
        ihdr_data = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
        ihdr = png_chunk(b'IHDR', ihdr_data)
        
        raw_data = b''
        for y in range(height):
            raw_data += b'\x00'
            for x in range(width):
                idx = y * width + x
                r, g, b = pixels[idx]
                raw_data += bytes([r, g, b])
        
        compressed = zlib.compress(raw_data, 9)
        idat = png_chunk(b'IDAT', compressed)
        iend = png_chunk(b'IEND', b'')
        
        return signature + ihdr + idat + iend

def hex_to_rgb(hex_str):
    hex_str = hex_str.lstrip('#')
    return tuple(int(hex_str[i:i+2], 16) for i in (0, 2, 4))

def draw_icon_pil(size):
    img = Image.new('RGB', (size, size), (255, 255, 255))
    draw = ImageDraw.Draw(img)
    
    scale = size / 128.0
    
    def px(v):
        return int(v * scale)
    
    cx, cy = px(64), px(64)
    r = px(56)
    
    from PIL import Image, ImageDraw
    gradient_img = Image.new('RGB', (size, size), (255, 255, 255))
    for y in range(size):
        for x in range(size):
            dx = x - cx
            dy = y - cy
            dist = (dx*dx + dy*dy)**0.5
            if dist <= r:
                t = y / (size - 1)
                r_val = int(66 + (52 - 66) * t)
                g_val = int(133 + (168 - 133) * t)
                b_val = int(244 + (83 - 244) * t)
                gradient_img.putpixel((x, y), (r_val, g_val, b_val))
    
    draw.bitmap((0, 0), gradient_img, fill=None)
    
    bw, bh = px(80), px(60)
    bx, by = px(24), px(28)
    draw.rounded_rectangle([bx, by, bx + bw, by + bh], radius=px(6), fill=(255, 255, 255, 242))
    
    header_h = px(16)
    draw.rounded_rectangle([bx, by, bx + bw, by + header_h], radius=px(6), fill=(232, 234, 237))
    draw.rectangle([bx, by + px(8), bx + bw, by + px(16)], fill=(232, 234, 237))
    
    dot_r = px(3)
    dot_y = by + px(8)
    colors = [(234, 67, 53), (251, 188, 5), (52, 168, 83)]
    for i, color in enumerate(colors):
        dot_x = bx + px(8) + i * px(10)
        draw.ellipse([dot_x - dot_r, dot_y - dot_r, dot_x + dot_r, dot_y + dot_r], fill=color)
    
    url_x, url_y = bx + px(38), by + px(4)
    url_w, url_h = px(36), px(8)
    draw.rounded_rectangle([url_x, url_y, url_x + url_w, url_y + url_h], radius=px(4), fill=(255, 255, 255))
    
    dom_x, dom_y = px(44), px(52)
    dom_size = px(32)
    
    stroke_w = px(3)
    draw.line([dom_x, dom_y + px(8), dom_x, dom_y], fill=(66, 133, 244), width=stroke_w)
    draw.line([dom_x, dom_y, dom_x + px(8), dom_y], fill=(66, 133, 244), width=stroke_w)
    
    draw.line([dom_x + dom_size, dom_y + px(8), dom_x + dom_size, dom_y], fill=(66, 133, 244), width=stroke_w)
    draw.line([dom_x + dom_size, dom_y, dom_x + dom_size - px(8), dom_y], fill=(66, 133, 244), width=stroke_w)
    
    draw.line([dom_x, dom_y + px(24), dom_x, dom_y + px(32)], fill=(66, 133, 244), width=stroke_w)
    draw.line([dom_x, dom_y + px(32), dom_x + px(8), dom_y + px(32)], fill=(66, 133, 244), width=stroke_w)
    
    draw.line([dom_x + dom_size, dom_y + px(24), dom_x + dom_size, dom_y + px(32)], fill=(66, 133, 244), width=stroke_w)
    draw.line([dom_x + dom_size, dom_y + px(32), dom_x + dom_size - px(8), dom_y + px(32)], fill=(66, 133, 244), width=stroke_w)
    
    bar_colors = [(52, 168, 83), (251, 188, 5), (234, 67, 53)]
    bar_heights = [px(12), px(16), px(8)]
    bar_y = dom_y + px(8)
    for i, (color, bar_w) in enumerate(zip(bar_colors, bar_heights)):
        bar_x = dom_x + px(10)
        draw.rounded_rectangle([bar_x, bar_y + i * px(6), bar_x + bar_w, bar_y + i * px(6) + px(3)], radius=px(1.5), fill=color)
    
    g_circle_x, g_circle_y = px(96), px(88)
    g_circle_r = px(8)
    draw.ellipse([g_circle_x - g_circle_r, g_circle_y - g_circle_r, g_circle_x + g_circle_r, g_circle_y + g_circle_r], fill=(0, 173, 216))
    
    return img

def draw_icon_pixel(size):
    pixels = [(255, 255, 255)] * (size * size)
    
    scale = size / 128.0
    def px(v):
        return int(v * scale)
    
    cx, cy = px(64), px(64)
    r = px(56)
    
    for y in range(size):
        for x in range(size):
            dx = x - cx
            dy = y - cy
            dist = (dx*dx + dy*dy)**0.5
            if dist <= r:
                t = y / (size - 1)
                r_val = int(66 + (52 - 66) * t)
                g_val = int(133 + (168 - 133) * t)
                b_val = int(244 + (83 - 244) * t)
                pixels[y * size + x] = (r_val, g_val, b_val)
    
    bw, bh = px(80), px(60)
    bx, by = px(24), px(28)
    radius = px(6)
    
    for y in range(by, by + bh):
        for x in range(bx, bx + bw):
            in_rect = True
            if x < bx + radius and y < by + radius:
                in_rect = (x - (bx + radius))**2 + (y - (by + radius))**2 <= radius**2
            elif x > bx + bw - radius and y < by + radius:
                in_rect = (x - (bx + bw - radius))**2 + (y - (by + radius))**2 <= radius**2
            elif x < bx + radius and y > by + bh - radius:
                in_rect = (x - (bx + radius))**2 + (y - (by + bh - radius))**2 <= radius**2
            elif x > bx + bw - radius and y > by + bh - radius:
                in_rect = (x - (bx + bw - radius))**2 + (y - (by + bh - radius))**2 <= radius**2
            if in_rect:
                pixels[y * size + x] = (255, 255, 255)
    
    header_h = px(16)
    for y in range(by, by + header_h):
        for x in range(bx, bx + bw):
            pixels[y * size + x] = (232, 234, 237)
    
    dot_r = px(3)
    dot_y = by + px(8)
    colors = [(234, 67, 53), (251, 188, 5), (52, 168, 83)]
    for i, color in enumerate(colors):
        dot_x = bx + px(8) + i * px(10)
        for dy in range(-dot_r, dot_r + 1):
            for dx in range(-dot_r, dot_r + 1):
                if dx*dx + dy*dy <= dot_r*dot_r:
                    ny, nx = dot_y + dy, dot_x + dx
                    if 0 <= ny < size and 0 <= nx < size:
                        pixels[ny * size + nx] = color
    
    url_x, url_y = bx + px(38), by + px(4)
    url_w, url_h = px(36), px(8)
    for y in range(url_y, url_y + url_h):
        for x in range(url_x, url_x + url_w):
            pixels[y * size + x] = (255, 255, 255)
    
    dom_x, dom_y = px(44), px(52)
    dom_size = px(32)
    stroke_w = px(3)
    
    def draw_line(x1, y1, x2, y2, color):
        dx = abs(x2 - x1)
        dy = abs(y2 - y1)
        sx = 1 if x1 < x2 else -1
        sy = 1 if y1 < y2 else -1
        err = dx - dy
        while x1 != x2 or y1 != y2:
            for dw in range(-(stroke_w // 2), stroke_w // 2 + 1):
                for dh in range(-(stroke_w // 2), stroke_w // 2 + 1):
                    nx, ny = x1 + dw, y1 + dh
                    if 0 <= nx < size and 0 <= ny < size:
                        pixels[ny * size + nx] = color
            e2 = 2 * err
            if e2 > -dy:
                err -= dy
                x1 += sx
            if e2 < dx:
                err += dx
                y1 += sy
    
    draw_line(dom_x, dom_y + px(8), dom_x, dom_y, (66, 133, 244))
    draw_line(dom_x, dom_y, dom_x + px(8), dom_y, (66, 133, 244))
    draw_line(dom_x + dom_size, dom_y + px(8), dom_x + dom_size, dom_y, (66, 133, 244))
    draw_line(dom_x + dom_size, dom_y, dom_x + dom_size - px(8), dom_y, (66, 133, 244))
    draw_line(dom_x, dom_y + px(24), dom_x, dom_y + px(32), (66, 133, 244))
    draw_line(dom_x, dom_y + px(32), dom_x + px(8), dom_y + px(32), (66, 133, 244))
    draw_line(dom_x + dom_size, dom_y + px(24), dom_x + dom_size, dom_y + px(32), (66, 133, 244))
    draw_line(dom_x + dom_size, dom_y + px(32), dom_x + dom_size - px(8), dom_y + px(32), (66, 133, 244))
    
    bar_colors = [(52, 168, 83), (251, 188, 5), (234, 67, 53)]
    bar_widths = [px(12), px(16), px(8)]
    bar_y_start = dom_y + px(8)
    for i, (color, bar_w) in enumerate(zip(bar_colors, bar_widths)):
        bar_x = dom_x + px(10)
        bar_y = bar_y_start + i * px(6)
        for y in range(bar_y, bar_y + px(3)):
            for x in range(bar_x, bar_x + bar_w):
                pixels[y * size + x] = color
    
    g_circle_x, g_circle_y = px(96), px(88)
    g_circle_r = px(8)
    for dy in range(-g_circle_r, g_circle_r + 1):
        for dx in range(-g_circle_r, g_circle_r + 1):
            if dx*dx + dy*dy <= g_circle_r*g_circle_r:
                ny, nx = g_circle_y + dy, g_circle_x + dx
                if 0 <= ny < size and 0 <= nx < size:
                    pixels[ny * size + nx] = (0, 173, 216)
    
    return pixels

output_dir = r"d:\Source\ileego\go_chrome\extension\icons"
os.makedirs(output_dir, exist_ok=True)

sizes = [16, 48, 128]
for size in sizes:
    filepath = os.path.join(output_dir, f"icon-{size}.png")
    if USE_PIL:
        img = draw_icon_pil(size)
        img.save(filepath)
    else:
        pixels = draw_icon_pixel(size)
        png_data = create_png(size, size, pixels)
        with open(filepath, "wb") as f:
            f.write(png_data)
    print(f"Created: {filepath}")

print("Done!")
