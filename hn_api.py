from datetime import datetime
import humanize
import requests

API_ENDPOINT = 'https://hacker-news.firebaseio.com/v0/'


def get_items_data(item_ids: list, get_comments: bool = True) -> list:
    items = []
    for item_id in item_ids:
        r = requests.get(f'{API_ENDPOINT}item/{item_id}.json')
        data = r.json()

        timestamp = datetime.fromtimestamp(data['time'])
        now = datetime.now()
        data['display_time'] = humanize.naturaldelta(now - timestamp)
        items.append(data)

    return items


def get_items_list(story_type: str, start_id: int, limit: int) -> dict:
    result = {'items': [], 'next_page_start_id': -1}

    url = f'{API_ENDPOINT}{story_type}stories.json'
    r = requests.get(url)

    story_ids = r.json()
    start_index = 0
    if start_id != -1:
        try:
            start_index = story_ids.index(start_id)
        except Exception:
            pass

    end_index = start_index + limit
    if start_index >= len(story_ids):
        start_index, end_index = 0, 0
    elif end_index >= len(story_ids):
        end_index = len(story_ids)

    result['items'] = get_items_data(story_ids[start_index:end_index])
    if end_index > 0 and end_index < len(story_ids):
        result['next_page_start_id'] = story_ids[end_index]

    return result
