# We sadly need this line due to a bug in the Jenkins Docker Build Step plugin.
FROM amsterdam/docker_python:latest

# Build the one Go binary we need for database import:
FROM golang:latest as gobuilder
WORKDIR /go/src/xmlparser
COPY *.go /go/src/xmlparser/
RUN go get
RUN go build

# Build the Python app and copy the Go binary to it:
FROM amsterdam/docker_python:latest

WORKDIR /app
COPY requirements.txt /app
RUN pip install --no-cache-dir -r requirements.txt

COPY . /app/
COPY --from=gobuilder /go/src/xmlparser/xmlparser /bin/xmlparser

## Do the .jenkins directory dance to enable data imports:
#COPY .jenkins/import /.jenkins-import/
#COPY .jenkins /app/.jenkins


#USER datapunt

