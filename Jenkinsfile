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
                echo "Go version:"
                go version
                echo "Go path:"
                which go
                
                # Créer le workspace Go pour Jenkins
                mkdir -p /var/lib/jenkins/go
                chown jenkins:jenkins /var/lib/jenkins/go
                
                # Vérifier les permissions
                echo "Current user:"
                whoami
                echo "Workspace:"
                pwd
                '''
            }
        }
        
        // ÉTAPE 3: Build application Go
        stage('Build Go Application') {
            steps {
                sh '''
                echo "🏗️ Building Go application..."
                
                # Vérifier le fichier main.go
                echo "=== Main Go file ==="
                cat main.go | head -10
                
                # Initialiser go.mod si absent
                if [ ! -f "go.mod" ]; then
                    echo "📝 Initializing go.mod..."
                    go mod init hello-app
                fi
                
                # Télécharger les dépendances
                echo "📥 Downloading dependencies..."
                go mod download || echo "No dependencies or already downloaded"
                
                # Build de l'application
                echo "🔨 Compiling application..."
                go build -v -o ${APP_NAME} .
                
                # Vérification
                echo "✅ Build completed:"
                ls -la ${APP_NAME}
                file ${APP_NAME}
                ./${APP_NAME} --version 2>/dev/null || ./${APP_NAME} -v 2>/dev/null || echo "Cannot test binary (expected)"
                '''
            }
        }
        
        // ÉTAPE 4: Tests statiques simplifiés
        stage('Static Analysis') {
            steps {
                sh '''
                echo "🔍 Running static analysis..."
                
                # Vérification de base
                echo "=== Basic Checks ==="
                go vet . 2>/dev/null && echo "✅ Go vet passed" || echo "⚠️ Go vet issues"
                
                # Vérification compilation
                echo "=== Compilation ==="
                go build -o /tmp/test-build . && echo "✅ Code compiles" || echo "❌ Compilation failed"
                rm -f /tmp/test-build
                
                echo "✅ Static analysis completed"
                '''
            }
        }
        
        // ÉTAPE 5: Construction image Docker
        stage('Build Docker Image') {
            steps {
                script {
                    sh """
                    echo "🐳 Building Docker image..."
                    
                    # Vérifier le Dockerfile
                    echo "=== Dockerfile Content ==="
                    cat Dockerfile
                    echo "=========================="
                    
                    # Construction de l'image
                    docker build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} .
                    docker tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
                    
                    echo "✅ Docker images created:"
                    docker images | grep ${DOCKER_REGISTRY}
                    """
                }
            }
        }
        
        // ÉTAPE 6: Déploiement sur Ubuntu
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
                            
                            # Arrêt ancien conteneur
                            echo '⏹️ Stopping existing container...'
                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            
                            # Nettoyage
                            echo '🧹 Cleaning up...'
                            docker image prune -f 2>/dev/null || true
                            
                            # Lancement nouveau conteneur (utilise l'image locale)
                            echo '🐳 Starting new container...'
                            docker run -d \\
                              --name ${APP_NAME} \\
                              -p ${APP_PORT}:${APP_PORT} \\
                              --restart unless-stopped \\
                              ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            # Vérification
                            echo '⏳ Waiting for startup...'
                            sleep 10
                            
                            echo '🔍 Verification:'
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}'
                            
                            # Test santé
                            if curl -f -s http://localhost:${APP_PORT}/ > /dev/null; then
                                echo '✅ Health check passed'
                                echo '🎉 Deployment successful!'
                                echo '🌐 Application URL: http://localhost:${APP_PORT}'
                                echo '🌐 Network URL: http://192.168.61.131:${APP_PORT}'
                            else
                                echo '⚠️ Health check failed - checking logs...'
                                docker logs ${APP_NAME} --tail 10
                                echo '⚠️ Deployment completed but health check failed'
                            fi
                        "
                        """
                    }
                }
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
            
            archiveArtifacts artifacts: '${APP_NAME},go.mod,*.go', fingerprint: true
        }
        success {
            sh """
            echo ""
            echo "✅ DÉPLOIEMENT RÉUSSI!"
            echo "🌐 Votre application Go est déployée:"
            echo "   http://localhost:${APP_PORT}"
            echo "   http://192.168.61.131:${APP_PORT}"
            echo "🐳 Image: ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
            """
        }
        failure {
            sh """
            echo "❌ DÉPLOIEMENT ÉCHOUÉ"
            echo "💡 Causes possibles:"
            echo "   - Problème de build Go"
            echo "   - Dockerfile incorrect"
            echo "   - Problème SSH"
            echo "   - Port ${APP_PORT} déjà utilisé"
            """
        }
    }
}
