# attack.star - Example Galick script
# This script demonstrates dynamic request generation.

def request():
    """
    This function is called for every request.
    It must return a dictionary with 'method', 'url', and optionally 'body'.
    """
    # You could use random data here if needed
    user_id = 123

    return {
        "method": "GET",
        "url": "https://httpbin.org/get?user_id={}".format(user_id),
        "headers": {
            "Content-Type": "application/json",
        }
    }
