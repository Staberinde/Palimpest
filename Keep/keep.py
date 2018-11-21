import os

from flask import Flask, Response
from flask_restplus import Api, Resource, Namespace
import gkeepapi

app = Flask(__name__)
api = Api(app, title='Keep Service API', version='1.0', prefix='/api/v1')


@api.route('/notes')
class Notes(Resource):
    keep = gkeepapi.Keep()

    def get(self):
        self.keep.login(
            os.environ['GOOGLE_EMAIL'],
            os.environ['GOOGLE_PASSWORD']
        )
        return Response(self.keep.all())
