#!/usr/bin/env python3
"""
示例搜索插件 - The Pirate Bay
这是一个演示如何编写 alist 搜索插件的示例
"""

import argparse
import json
import sys
import re
import requests
from urllib.parse import quote_plus, urljoin
from bs4 import BeautifulSoup


class TPBSearchPlugin:
    """The Pirate Bay 搜索插件"""

    def __init__(self):
        self.name = "thepiratebay"
        self.display_name = "The Pirate Bay"
        self.version = "1.0.0"
        self.base_url = "https://thepiratebay.org"
        self.categories = [
            "all", "audio", "video", "applications", "games", "other"
        ]

        # 分类映射
        self.category_map = {
            "all": "0",
            "audio": "100",
            "video": "200",
            "applications": "300",
            "games": "400",
            "other": "600"
        }

    def get_info(self):
        """返回插件信息"""
        return {
            "display_name": self.display_name,
            "version": self.version,
            "categories": self.categories
        }

    def search(self, query, category="all", page=1):
        """执行搜索"""
        try:
            # 构建搜索URL
            cat_id = self.category_map.get(category, "0")
            search_url = f"{self.base_url}/search/{quote_plus(query)}/1/99/{cat_id}"

            headers = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
            }

            response = requests.get(search_url, headers=headers, timeout=30)
            response.raise_for_status()

            return self.parse_results(response.text)

        except Exception as e:
            sys.stderr.write(f"Search error: {str(e)}\n")
            return []

    def parse_results(self, html):
        """解析搜索结果"""
        results = []

        try:
            soup = BeautifulSoup(html, 'html.parser')

            # 查找结果行
            for row in soup.find_all('tr')[1:]:  # 跳过标题行
                try:
                    cells = row.find_all('td')
                    if len(cells) < 4:
                        continue

                    # 提取标题和链接
                    title_cell = cells[1]
                    title_link = title_cell.find('a')
                    if not title_link:
                        continue

                    title = title_link.get_text(strip=True)
                    detail_url = urljoin(self.base_url, title_link.get('href', ''))

                    # 提取磁力链接
                    magnet_link = ""
                    magnet_elem = row.find('a', href=re.compile(r'^magnet:'))
                    if magnet_elem:
                        magnet_link = magnet_elem.get('href', '')

                    # 提取种子数和下载数
                    seeds = 0
                    leechs = 0
                    if len(cells) >= 3:
                        seeds_text = cells[-2].get_text(strip=True)
                        leechs_text = cells[-1].get_text(strip=True)

                        try:
                            seeds = int(seeds_text) if seeds_text.isdigit() else 0
                            leechs = int(leechs_text) if leechs_text.isdigit() else 0
                        except ValueError:
                            pass

                    # 提取文件大小
                    size = ""
                    desc_elem = title_cell.find('font', class_='detDesc')
                    if desc_elem:
                        desc_text = desc_elem.get_text()
                        size_match = re.search(r'Size (\d+\.?\d*\s*[KMGT]?iB)', desc_text)
                        if size_match:
                            size = size_match.group(1)

                    # 确定分类
                    category = "other"
                    category_cell = cells[0]
                    category_links = category_cell.find_all('a')
                    if category_links:
                        cat_text = category_links[0].get_text(strip=True).lower()
                        if 'audio' in cat_text or 'music' in cat_text:
                            category = "audio"
                        elif 'video' in cat_text or 'movie' in cat_text:
                            category = "video"
                        elif 'application' in cat_text or 'software' in cat_text:
                            category = "applications"
                        elif 'game' in cat_text:
                            category = "games"

                    result = {
                        "title": title,
                        "url": detail_url,
                        "torrent_url": "",  # TPB 通常不提供直接的种子下载
                        "magnet_link": magnet_link,
                        "size": size,
                        "seeds": seeds,
                        "leechs": leechs,
                        "category": category
                    }

                    results.append(result)

                except Exception as e:
                    sys.stderr.write(f"Error parsing result row: {str(e)}\n")
                    continue

        except Exception as e:
            sys.stderr.write(f"Error parsing HTML: {str(e)}\n")

        return results


def main():
    plugin = TPBSearchPlugin()

    parser = argparse.ArgumentParser(description=f'{plugin.display_name} Search Plugin')
    parser.add_argument('--info', action='store_true', help='Return plugin information')
    parser.add_argument('--search', help='Search query')
    parser.add_argument('--category', default='all', help='Search category')
    parser.add_argument('--page', type=int, default=1, help='Page number')

    args = parser.parse_args()

    if args.info:
        # 返回插件信息
        info = plugin.get_info()
        print(json.dumps(info, ensure_ascii=False, indent=2))
    elif args.search:
        # 执行搜索
        results = plugin.search(args.search, args.category, args.page)
        print(json.dumps(results, ensure_ascii=False, indent=2))
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()