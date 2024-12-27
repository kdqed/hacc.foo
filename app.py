from flask import (
    abort,
    Flask,
    request,
    render_template,
)
import hn_api

app = Flask(__name__, template_folder='templates/')


def wrap(html_content):
    return render_template('_wrapper.html', html_content=html_content)


@app.route('/')
@app.route('/top')
@app.route('/new')
@app.route('/best')
@app.route('/ask')
@app.route('/show')
@app.route('/job')
def item_list_page():
    story_type = request.path.strip("/")
    if story_type == '':
        story_type = 'top'

    start_id = int(request.args.get('start_id', -1))
    result = hn_api.get_items_list(story_type, start_id, 5)

    response = render_template(
        'list.html',
        result=result,
        story_type=story_type,
    )

    if request.headers.get('HX-Request') == 'true':
        return response
    else:
        return wrap(response)


@app.route('/item/<item_id>')
def item_page(item_id):
    item_id = int(item_id)
    result = hn_api.get_items_data([item_id], get_comments=True)
    if not len(result):
        return abort(404)

    item = result[0]
    item_html = render_template('item.html', item=item)
    if request.headers.get('HX-Request') == 'true':
        return item_html
    else:
        return wrap(item_html)


if __name__ == "__main__":
    app.run(host='127.0.0.1', port='2511', debug=True, use_reloader=True)
