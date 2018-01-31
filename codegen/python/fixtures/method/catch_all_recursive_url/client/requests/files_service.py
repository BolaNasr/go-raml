# DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
from .File import File
from .unhandled_api_error import UnhandledAPIError
from .unmarshall_error import UnmarshallError


class FilesService:
    def __init__(self, client):
        self.client = client

    def files_byPath_get(self, path, headers=None, query_params=None, content_type="application/json"):
        """
        It is method for GET /files/{path:*}
        """
        uri = self.client.base_url + "/files" + path
        resp = self.client.get(uri, None, headers, query_params, content_type)
        try:
            if resp.status_code == 200:
                return File(resp.json()), resp

            message = 'unknown status code={}'.format(resp.status_code)
            raise UnhandledAPIError(response=resp, code=resp.status_code,
                                    message=message)
        except ValueError as msg:
            raise UnmarshallError(resp, msg)
        except UnhandledAPIError as uae:
            raise uae
        except Exception as e:
            raise UnmarshallError(resp, e.message)
