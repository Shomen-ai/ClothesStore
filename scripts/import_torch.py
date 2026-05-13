#!/usr/bin/env python3
"""
Import torch-fff.com catalogue into local ClothesStore.

Usage:
  python import_torch.py scrape          # parse site -> products.json + _images/
  python import_torch.py upload          # push parsed data to local admin API
  python import_torch.py all             # scrape + upload
"""
import json
import os
import re
import sys
import time
from pathlib import Path
from urllib.parse import urljoin

import requests
from bs4 import BeautifulSoup

BASE = "https://torch-fff.com"
UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36"
COOKIES = {"beget": "begetok"}

API = os.environ.get("API_URL", "http://localhost:8080/api")
ADMIN_EMAIL = os.environ.get("ADMIN_EMAIL", "admin@clothesstore.ru")
ADMIN_PASSWORD = os.environ.get("ADMIN_PASSWORD", "admin123")

HERE = Path(__file__).resolve().parent
DATA_FILE = HERE / "products.json"
IMG_DIR = HERE / "_images"

MAX_IMAGES = int(os.environ.get("TORCH_MAX_IMAGES", "8"))


def get(url, **kw):
    last = None
    for attempt in range(4):
        try:
            r = requests.get(url, headers={"User-Agent": UA}, cookies=COOKIES, timeout=30, **kw)
            r.raise_for_status()
            return r
        except (requests.exceptions.SSLError, requests.exceptions.ConnectionError, requests.exceptions.Timeout) as e:
            last = e
            time.sleep(1 + attempt)
    raise last


def fetch_categories():
    """Map PROP_6 id -> category name from /shop/ filters."""
    html = get(f"{BASE}/shop/").text
    cats = {}
    for m in re.finditer(
        r'class="filter__category-btn filter__category-btn-type[^"]*"\s*value="(\d+)"\s*name="PROP_6">\s*([^<]+?)\s*</button>',
        html,
    ):
        cats[m.group(1)] = re.sub(r"\s+", " ", m.group(2)).strip()
    return cats


def fetch_category_products(prop6):
    """Return list of product slugs (relative URLs) for a given category."""
    slugs = []
    seen = set()
    for page in range(1, 10):
        url = f"{BASE}/shop/?PROP_6={prop6}&PAGEN_1={page}"
        html = get(url).text
        page_slugs = set(re.findall(r'href="(/shop/[^"/]+/)"', html))
        page_slugs.discard("/shop/")
        new = page_slugs - seen
        if not new:
            break
        for s in new:
            slugs.append(s)
            seen.add(s)
    return slugs


def fetch_sale_slugs():
    """All product slugs that are currently on SALE (PROP_5=on)."""
    slugs = set()
    for page in range(1, 10):
        url = f"{BASE}/shop/?PROP_5=on&PAGEN_1={page}"
        html = get(url).text
        page_slugs = set(re.findall(r'href="(/shop/[^"/]+/)"', html))
        page_slugs.discard("/shop/")
        if page_slugs.issubset(slugs):
            break
        slugs |= page_slugs
    return slugs


def parse_product(slug):
    """Parse a single product page into a dict."""
    url = urljoin(BASE, slug)
    html = get(url).text
    soup = BeautifulSoup(html, "html.parser")

    # Name: last breadcrumb <span class="page">
    name = None
    for span in soup.select("span.page"):
        txt = span.get_text(strip=True)
        if txt and txt.upper() != "КАТАЛОГ":
            name = txt
    if not name:
        title = soup.find("title")
        name = title.get_text(strip=True).replace("Torch - ", "") if title else slug

    # Price (current)
    price = 0.0
    pm = soup.find(attrs={"itemprop": "price"})
    if pm and pm.get("content"):
        price = float(pm["content"])

    # Old price -> is_on_sale (filled if discounted)
    is_on_sale = False
    old = soup.find(id=re.compile(r"_old_price$"))
    if old and old.get_text(strip=True):
        is_on_sale = True

    # Sizes — distinct, each with stock; treat "9+" as 10
    sizes = []
    seen_sizes = set()
    for offer in soup.find_all(attrs={"itemprop": "offers"}):
        sku = offer.find("meta", attrs={"itemprop": "sku"})
        if not sku:
            continue
        s = sku.get("content", "").strip()
        if s and s not in seen_sizes:
            seen_sizes.add(s)
            sizes.append({"size": s, "stock_qty": 10})

    # Description from `.product__description-text > p`
    description = ""
    desc_el = soup.select_one("li.product__description-text p")
    if desc_el:
        for br in desc_el.find_all("br"):
            br.replace_with("\n")
        description = desc_el.get_text().strip()
        description = re.sub(r"\n{3,}", "\n\n", description)
        description = clean_description(description)

    # Images from main product slider (NOT recommend section, NOT size-table icons)
    images = []
    slider = soup.select_one('[data-entity="images-container"]')

    def is_product_img(url):
        u = url.lower()
        if "/upload/iblock/" not in u:
            return False
        if not u.endswith((".jpg", ".jpeg", ".png", ".webp")):
            return False
        return True

    if slider:
        for img in slider.find_all("img"):
            # Skip size table icon
            cls = " ".join(img.get("class") or [])
            if "table-size__img" in cls:
                continue
            src = img.get("data-src") or img.get("src")
            if src and is_product_img(src):
                images.append(urljoin(BASE, src))
        for a in slider.find_all("a"):
            href = a.get("href") or ""
            if is_product_img(href):
                images.append(urljoin(BASE, href))
    if not images:
        recommend_idx = html.find('product-recommend__title')
        scope = html[:recommend_idx] if recommend_idx > 0 else html
        for m in re.finditer(r'/upload/iblock/[a-f0-9]+/[A-Za-z0-9_-]+\.(?:jpg|jpeg|png|webp)', scope):
            images.append(urljoin(BASE, m.group(0)))
    # dedupe, preserve order, cap to MAX_IMAGES
    seen = set()
    images = [x for x in images if not (x in seen or seen.add(x))]
    images = images[:MAX_IMAGES]

    return {
        "slug": slug.strip("/").split("/")[-1],
        "name": name,
        "price": price,
        "is_on_sale": is_on_sale,
        "sizes": sizes,
        "description": description,
        "images": images,
    }


def clean_description(text):
    """Strip the manager-contact preamble that every torch-fff product carries."""
    lines = text.split("\n")
    skip_markers = ("менеджер", "вк:", "вк ", "tg:", "tg ", "телеграм", "vk.cc", "instagram", "@man_torch")
    out, started = [], False
    for ln in lines:
        low = ln.lower().strip()
        if not started:
            if not low:
                continue
            if any(m in low for m in skip_markers):
                continue
            started = True
        out.append(ln)
    cleaned = "\n".join(out).strip()
    return re.sub(r"\n{3,}", "\n\n", cleaned)


def slugify_cat(name):
    s = name.lower()
    s = re.sub(r"[^\w\s-]", "", s, flags=re.UNICODE)
    s = re.sub(r"\s+", "-", s)
    return s.strip("-")


def scrape():
    IMG_DIR.mkdir(exist_ok=True)
    print("Loading categories...")
    cats = fetch_categories()
    print(f"  {len(cats)} categories")

    print("Loading SALE list...")
    sale_set = fetch_sale_slugs()
    print(f"  {len(sale_set)} sale items")

    products = {}  # slug -> product dict (deduped)
    for prop6, cat_name in cats.items():
        print(f"\n[{cat_name}] (PROP_6={prop6})")
        slugs = fetch_category_products(prop6)
        print(f"  {len(slugs)} products")
        for s in slugs:
            try:
                p = parse_product(s)
            except Exception as e:
                print(f"  ! failed {s}: {e}")
                continue
            p["category"] = cat_name
            p["category_slug"] = slugify_cat(cat_name)
            if s in sale_set:
                p["is_on_sale"] = True
            if p["slug"] in products:
                continue
            products[p["slug"]] = p
            print(f"  + {p['slug']} | {p['name'][:40]} | {p['price']:.0f}₽ "
                  f"| {len(p['sizes'])} sz | {len(p['images'])} img"
                  f"{' | SALE' if p['is_on_sale'] else ''}")
            # download images
            pdir = IMG_DIR / p["slug"]
            pdir.mkdir(exist_ok=True)
            for i, url in enumerate(p["images"]):
                ext = os.path.splitext(url.split("?")[0])[1].lower() or ".jpg"
                target = pdir / f"{i:02d}{ext}"
                if target.exists() and target.stat().st_size > 1024:
                    continue
                try:
                    r = get(url)
                    target.write_bytes(r.content)
                except Exception as e:
                    print(f"    ! img fail {url}: {e}")
            time.sleep(0.2)
        time.sleep(0.5)

    DATA_FILE.write_text(
        json.dumps(list(products.values()), ensure_ascii=False, indent=2)
    )
    print(f"\nSaved {len(products)} products to {DATA_FILE}")


def admin_token():
    r = requests.post(f"{API}/auth/login",
                      json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD},
                      timeout=10)
    r.raise_for_status()
    body = r.json()
    return body.get("tokens", body)["access_token"]


def upload():
    if not DATA_FILE.exists():
        print(f"Run `scrape` first — no {DATA_FILE}")
        sys.exit(1)
    products = json.loads(DATA_FILE.read_text())
    token = admin_token()
    H = {"Authorization": f"Bearer {token}"}

    # Categories: collect unique, fetch existing, create missing
    existing = requests.get(f"{API}/admin/categories", headers=H, timeout=10).json() or []
    cat_by_slug = {c["slug"]: c["id"] for c in existing}
    needed = {(p["category"], p["category_slug"]) for p in products}
    for name, cslug in sorted(needed):
        if cslug in cat_by_slug:
            continue
        r = requests.post(f"{API}/admin/categories", headers=H,
                          json={"name": name, "slug": cslug, "sort_order": 0}, timeout=10)
        if r.status_code >= 300:
            print(f"  category create failed {name}: {r.status_code} {r.text}")
            continue
        cat_by_slug[cslug] = r.json()["id"]
        print(f"  + category {name} (id={cat_by_slug[cslug]})")

    # Existing products — skip by name (idempotent re-run)
    existing_products = requests.get(f"{API}/admin/products?page=1", headers=H, timeout=15).json() or []
    have_names = {p["name"] for p in existing_products}

    for p in products:
        if p["name"] in have_names:
            print(f"  = exists {p['name']}")
            continue
        cat_id = cat_by_slug.get(p["category_slug"])
        if not cat_id:
            print(f"  ! no category for {p['name']}")
            continue
        body = {
            "name": p["name"],
            "description": p["description"],
            "category_id": cat_id,
            "price": p["price"],
            "is_active": True,
            "is_on_sale": p["is_on_sale"],
            "sizes": p["sizes"],
        }
        r = requests.post(f"{API}/admin/products", headers=H, json=body, timeout=15)
        if r.status_code >= 300:
            print(f"  ! create failed {p['name']}: {r.status_code} {r.text}")
            continue
        pid = r.json()["id"]
        print(f"  + product {pid} | {p['name']}")

        # Upload images
        pdir = IMG_DIR / p["slug"]
        if not pdir.exists():
            continue
        files = []
        opened = []
        for fp in sorted(pdir.iterdir()):
            if fp.suffix.lower() not in (".jpg", ".jpeg", ".png", ".webp"):
                continue
            f = open(fp, "rb")
            opened.append(f)
            files.append(("images", (fp.name, f, "image/jpeg")))
        if files:
            ru = requests.post(f"{API}/admin/products/{pid}/images",
                               headers=H, files=files, timeout=120)
            if ru.status_code >= 300:
                print(f"    ! upload images failed: {ru.status_code} {ru.text[:200]}")
            else:
                print(f"    + {len(files)} images")
        for f in opened:
            f.close()


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else "scrape"
    if cmd == "scrape":
        scrape()
    elif cmd == "upload":
        upload()
    elif cmd == "all":
        scrape()
        upload()
    else:
        print(__doc__)
        sys.exit(1)


if __name__ == "__main__":
    main()
