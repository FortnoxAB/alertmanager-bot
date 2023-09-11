node('go1.21') {
	container('run'){
		def newTag = ''
		def tag = ''
		def gitTag = ''

		try {
			stage('Checkout'){
					checkout scm
					gitTag = sh(script: 'git tag -l --contains HEAD', returnStdout: true).trim()
			}


			stage('Fetch dependencies'){
				sh('go mod download')
			}
			stage('Run test'){
				sh('go test -v ./...')
			}

			if(gitTag != ''){
				tag = gitTag
			}

			if( tag != ''){
				strippedTag = tag.replaceFirst('v', '')
				stage('Build the application'){
					echo "Building with docker tag ${strippedTag}"
					sh('CGO_ENABLED=0 GOOS=linux go build')
				}

				stage('Generate docker image'){
					image = docker.build('fortnox/alertmanager-bot:'+strippedTag, '--pull .')
				}

				stage('Push docker image'){
					docker.withRegistry("https://quay.io", 'docker-registry') {
						image.push()
					}
				}
			}

		} catch(err) {
			throw err
		}
	}
}

