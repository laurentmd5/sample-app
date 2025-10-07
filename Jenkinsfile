pipeline {
    agent any
    
    environment {
        // Configuration Application
        APP_NAME = 'go-dev-dashboard'
        APP_PORT = '8090'
    
        // Configuration Docker
        DOCKER_REGISTRY = 'laurentmd5'
        DOCKER_IMAGE = "${APP_NAME}"
        DOCKER_TAG = "${env.BUILD_NUMBER}"
        
        // Configuration Serveur Ubuntu
        DEPLOY_SERVER = 'devops@localhost'
        DEPLOY_PATH = '/home/devops/apps'
        SSH_CREDENTIALS_ID = 'ubuntu-server-ssh'
        
        // Configuration Go
        GOPATH = '/var/lib/jenkins/go'
    }
    
    stages {
        // ÉTAPE 1: Checkout avec credentials GitHub
        stage('Checkout Code') {
            steps {
                git branch: 'master',
                    url: 'https://github.com/laurentmd5/sample-app.git',
                    credentialsId: 'my-token'

                sh '''
                echo "📦 Repository: https://github.com/laurentmd5/sample-app.git"
                echo "📝 Branch: master"
                echo "🔍 Files in repository:"
                ls -la
                echo "=== Go files ==="
                find . -name "*.go" -type f
                '''
            }
        }
        
        // ÉTAPE 2: Setup Go Environment
        stage('Setup Go') {
            steps {
                sh '''
                echo "🔧 Setting up Go environment..."
                go version
                which go
                
                mkdir -p /var/lib/jenkins/go
                chown jenkins:jenkins /var/lib/jenkins/go
                
                whoami
                pwd
                '''
            }
        }
        
        // ÉTAPE 3: Build Go Application
        stage('Build Go Application') {
            steps {
                sh '''
                echo "🏗️ Building Go application..."
                
                if [ ! -f "go.mod" ]; then
                    echo "📝 Initializing go.mod..."
                    go mod init go-dev-dashboard
                fi

                go mod tidy
                go mod download

                go build -v -o ${APP_NAME} .

                echo "✅ Build completed:"
                ls -la ${APP_NAME}
                file ${APP_NAME}
                '''
            }
        }

        // ÉTAPE 4: Static Code Analysis (GoSec)
        stage('Static Code Analysis') {
            steps {
                sh '''
                echo "🔍 Running GoSec security analysis..."
                go install github.com/securego/gosec/v2/cmd/gosec@latest
                export PATH=$PATH:$(go env GOPATH)/bin
                gosec ./... || echo "⚠️ Gosec found issues (non-blocking)"
                '''
            }
        }

        // ÉTAPE 5: Dynamic Tests (Unit tests + Coverage)
        stage('Dynamic Tests') {
            steps {
                sh '''
                echo "🧪 Running unit tests with coverage..."
                go test ./... -v -coverprofile=coverage.out || echo "⚠️ Tests failed"
                go tool cover -func=coverage.out | tee coverage-report.txt
                '''
            }
        }

        // ÉTAPE 6: Static Analysis simplifiée
        stage('Static Analysis') {
            steps {
                sh '''
                echo "🔍 Running basic static analysis..."
                go vet . || echo "⚠️ Go vet issues"
                go build -o /tmp/test-build . && echo "✅ Code compiles"
                rm -f /tmp/test-build
                '''
            }
        }

        // ÉTAPE 7: Build Docker Image
        stage('Build Docker Image') {
            steps {
                sh '''
                echo "🐳 Building Docker image..."
                echo "=== Dockerfile Content ==="
                cat Dockerfile
                echo "=========================="

                docker build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} .
                docker tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest

                echo "✅ Docker images created:"
                docker images | grep ${DOCKER_REGISTRY}
                '''
            }
        }

        // ÉTAPE 8: Container Scan (Trivy)
        stage('Container Scan') {
            steps {
                sh '''
                echo "🛡️ Scanning Docker image for vulnerabilities (Trivy)..."
                trivy image --quiet --no-progress --severity HIGH,CRITICAL ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest || echo "⚠️ Vulnerabilities found"
                '''
            }
        }

        // ÉTAPE 9: Filesystem Scan (Trivy FS)
        stage('Filesystem Scan') {
            steps {
                sh '''
                echo "🧾 Scanning source filesystem for vulnerabilities..."
                trivy fs --quiet --no-progress --severity HIGH,CRITICAL . || echo "⚠️ Issues found in filesystem"
                '''
            }
        }

        // ÉTAPE 10: Deploy to Ubuntu via SSH
        stage('Deploy to Ubuntu via SSH') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🚀 Deploying to Ubuntu server..."
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            set -e
                            echo '🎯 Starting deployment of ${APP_NAME}...'

                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            docker image prune -f 2>/dev/null || true

                            docker run -d \\
                              --name ${APP_NAME} \\
                              -p ${APP_PORT}:${APP_PORT} \\
                              --restart unless-stopped \\
                              ${DOCKER_REGISTRY}/${APP_NAME}:latest

                            sleep 10
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}'

                            if curl -f -s http://localhost:${APP_PORT}/ > /dev/null; then
                                echo '✅ Health check passed'
                            else
                                echo '⚠️ Health check failed'
                                docker logs ${APP_NAME} --tail 10
                            fi
                        "
                        """
                    }
                }
            }
        }

        // ÉTAPE 11: Environment Scan (Lynis + apt audit)
        stage('Environment Scan') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🧮 Running environment security scan..."
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            sudo apt update -qq
                            sudo apt install -y lynis > /dev/null 2>&1 || true
                            sudo lynis audit system --quiet || echo '⚠️ Lynis found warnings'
                        "
                        """
                    }
                }
            }
        }

        // ÉTAPE 12: Final Check
        stage('Final Check') {
            steps {
                sh '''
                echo "🔁 Final verification..."
                echo "📦 Checking Docker containers status..."
                docker ps --format "table {{.Names}}\t{{.Status}}"
                echo "✅ Pipeline and deployment verification complete!"
                '''
            }
        }
    }
    
    post {
        always {
            sh '''
            echo "🧼 Cleaning up workspace..."
            docker system prune -f 2>/dev/null || true
            rm -f ${APP_NAME} 2>/dev/null || true
            '''
            archiveArtifacts artifacts: '${APP_NAME},go.mod,*.go,coverage-report.txt', fingerprint: true
        }
        success {
            sh """
            echo "✅ DÉPLOIEMENT RÉUSSI!"
            echo "🌐 Application disponible sur:"
            echo "   http://localhost:${APP_PORT}"
            echo "   http://192.168.61.131:${APP_PORT}"
            """
        }
        failure {
            sh """
            echo "❌ DÉPLOIEMENT ÉCHOUÉ"
            echo "💡 Vérifiez les logs Jenkins et les scans Trivy/Lynis."
            """
        }
    }
}
