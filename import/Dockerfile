# We sadly need this line due to a bug in the Jenkins Docker Build Step plugin.
FROM amsterdam/python

# Build the one Go binary we need for database import:
FROM golang:latest as gobuilder
WORKDIR /go/src/importer
COPY *.go /go/src/importer/
RUN go get
RUN go build

# Build the Python app and copy the Go binary to it:
FROM amsterdam/python

WORKDIR /app
COPY requirements.txt /app
RUN pip install --no-cache-dir -r requirements.txt

COPY . /app/
COPY --from=gobuilder /go/src/importer/importer /bin/importer

## Do the .jenkins directory dance to enable data imports:
#COPY .jenkins/import /.jenkins-import/
#COPY .jenkins /app/.jenkins


#USER datapunt

