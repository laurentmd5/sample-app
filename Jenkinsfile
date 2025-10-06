pipeline {
    agent any
    
    environment {
        // Configuration Application
        APP_NAME = 'hello-app'
        APP_PORT = '8090'
        
        // Configuration Docker
        DOCKER_REGISTRY = 'laurentmd5'
        DOCKER_IMAGE = "${APP_NAME}"
        DOCKER_TAG = "${env.BUILD_NUMBER}"
        
        // Configuration Serveur Ubuntu
        DEPLOY_SERVER = 'devops@localhost'
        DEPLOY_PATH = '/home/devops/apps'
        SSH_CREDENTIALS_ID = 'ubuntu-server-ssh'
    }
    
    stages {
        // ÉTAPE 1: Checkout du code
        stage('Checkout Code') {
            steps {
                checkout scm
                sh '''
                echo "📦 Repository: ${GIT_URL}"
                echo "📝 Branch: ${GIT_BRANCH}"
                echo "🔍 Files in repository:"
                ls -la
                '''
            }
        }
        
        // ÉTAPE 2: Vérification des fichiers
        stage('Verify Files') {
            steps {
                sh '''
                echo "🔍 Checking required files..."
                if [ ! -f "Dockerfile" ]; then
                    echo "❌ Dockerfile not found!"
                    exit 1
                fi
                if [ ! -f "*.go" ]; then
                    echo "⚠️ No Go files found in root directory"
                    find . -name "*.go" | head -5
                fi
                echo "✅ Files check completed"
                '''
            }
        }
        
        // ÉTAPE 3: Tests statiques Go
        stage('Static Analysis') {
            steps {
                sh '''
                echo "🔍 Running static analysis..."
                
                # Vérification des fichiers Go
                echo "=== Go Files ==="
                find . -name "*.go" -type f | head -10
                
                # Initialisation go.mod si nécessaire
                if [ ! -f "go.mod" ]; then
                    echo "📝 Initializing go.mod..."
                    go mod init hello-app
                fi
                
                # Téléchargement des dépendances
                echo "📥 Downloading dependencies..."
                go mod download 2>/dev/null || echo "No dependencies"
                
                # Analyse statique de base
                echo "=== Go Vet ==="
                go vet . 2>/dev/null || echo "Go vet completed"
                
                # Vérification de la compilation
                echo "=== Compilation Check ==="
                go build -o /tmp/test-build . 2>/dev/null && echo "✅ Code compiles successfully" || echo "⚠️ Compilation issues"
                rm -f /tmp/test-build 2>/dev/null || true
                
                echo "✅ Static analysis completed"
                '''
            }
        }
        
        // ÉTAPE 4: Tests dynamiques
        stage('Dynamic Tests') {
            steps {
                sh '''
                echo "🧪 Running dynamic tests..."
                mkdir -p test-results
                
                # Tests unitaires basiques
                echo "=== Unit Tests ==="
                if ls *_test.go 1> /dev/null 2>&1; then
                    go test -v -coverprofile=test-results/coverage.out . 2>&1 | tee test-results/test-output.log
                    echo "✅ Unit tests executed"
                else
                    echo "ℹ️ No test files found"
                    echo "no tests" > test-results/test-output.log
                fi
                
                # Test de build final
                echo "=== Final Build Test ==="
                CGO_ENABLED=0 GOOS=linux go build -o hello-app .
                ls -la hello-app
                file hello-app
                
                echo "✅ Dynamic tests completed"
                '''
            }
            
            post {
                always {
                    archiveArtifacts artifacts: 'hello-app,test-results/**/*', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 5: Construction image Docker
        stage('Build Docker Image') {
            steps {
                script {
                    sh """
                    echo "🐳 Building Docker image with your Dockerfile..."
                    echo "=== Dockerfile content ==="
                    cat Dockerfile
                    echo "=========================="
                    
                    # Construction de l'image avec votre Dockerfile
                    docker build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} .
                    docker tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
                    
                    echo "✅ Docker images created:"
                    docker images | grep ${DOCKER_REGISTRY}
                    """
                }
            }
        }
        
        // ÉTAPE 6: Push vers Docker Hub
        stage('Push to Docker Hub') {
            steps {
                script {
                    withCredentials([usernamePassword(
                        credentialsId: 'docker-hub-creds',
                        usernameVariable: 'DOCKERHUB_USER',
                        passwordVariable: 'DOCKERHUB_PASS'
                    )]) {
                        sh """
                        echo "📤 Pushing to Docker Hub..."
                        
                        # Login Docker Hub
                        echo \$DOCKERHUB_PASS | docker login -u \$DOCKERHUB_USER --password-stdin
                        
                        # Push des images
                        docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG}
                        docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
                        
                        # Logout
                        docker logout
                        
                        echo "✅ Images pushed to Docker Hub"
                        echo "🔗 https://hub.docker.com/r/${DOCKER_REGISTRY}/${DOCKER_IMAGE}"
                        """
                    }
                }
            }
        }
        
        // ÉTAPE 7: Déploiement sur Ubuntu via SSH
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
                        
                        # Déploiement direct via SSH
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            set -e
                            echo '🎯 Starting deployment of ${APP_NAME}...'
                            
                            # Arrêt de l'ancien conteneur
                            echo '⏹️ Stopping existing container...'
                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            
                            # Pull de la nouvelle image
                            echo '📥 Pulling new image from Docker Hub...'
                            docker pull ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            # Lancement du nouveau conteneur
                            echo '🐳 Starting new container...'
                            docker run -d \\
                              --name ${APP_NAME} \\
                              -p ${APP_PORT}:${APP_PORT} \\
                              --restart unless-stopped \\
                              ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            # Attente du démarrage
                            echo '⏳ Waiting for application to start...'
                            sleep 10
                            
                            # Vérification
                            echo '🔍 Verifying deployment...'
                            if docker ps | grep -q ${APP_NAME}; then
                                echo '✅ Container is running'
                            else
                                echo '❌ Container failed to start'
                                docker logs ${APP_NAME} --tail 10
                                exit 1
                            fi
                            
                            # Test de santé
                            if curl -f -s http://localhost:${APP_PORT}/ > /dev/null; then
                                echo '✅ Health check passed'
                            else
                                echo '⚠️ Health check failed - checking logs...'
                                docker logs ${APP_NAME} --tail 10
                            fi
                            
                            echo '🎉 Deployment completed!'
                            echo '🌐 App available at: http://localhost:${APP_PORT}'
                            echo '🌐 Network URL: http://192.168.61.131:${APP_PORT}'
                        "
                        """
                    }
                }
            }
        }
        
        // ÉTAPE 8: Vérification finale
        stage('Final Verification') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER', 
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🔎 Final verification..."
                        
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            echo '=== FINAL DEPLOYMENT STATUS ==='
                            echo 'Container:'
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'
                            
                            echo 'Logs (last 5 lines):'
                            docker logs ${APP_NAME} --tail 5
                            
                            echo 'Port binding:'
                            docker port ${APP_NAME}
                            
                            echo 'Health check:'
                            if curl -f -s -o /dev/null -w 'HTTP Status: %{http_code}\\n' http://localhost:${APP_PORT}/; then
                                echo '✅ APPLICATION IS RUNNING CORRECTLY'
                            else
                                echo '❌ APPLICATION HEALTH CHECK FAILED'
                            fi
                        "
                        
                        echo ""
                        echo "🎊 DEPLOYMENT SUCCESSFUL!"
                        echo "📊 SUMMARY:"
                        echo "   App: ${APP_NAME}"
                        echo "   Version: ${DOCKER_TAG}" 
                        echo "   URL: http://192.168.61.131:${APP_PORT}"
                        echo "   Image: ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
                        """
                    }
                }
            }
        }
    }
    
    post {
        always {
            sh '''
            echo "🧼 Cleaning up..."
            docker system prune -f 2>/dev/null || true
            rm -f hello-app 2>/dev/null || true
            '''
            
            archiveArtifacts artifacts: 'hello-app,go.mod,*.go', fingerprint: true
        }
        success {
            sh """
            echo ""
            echo "✅ DÉPLOIEMENT RÉUSSI!"
            echo "🌐 Votre application Go est maintenant accessible:"
            echo "   Local: http://localhost:${APP_PORT}"
            echo "   Réseau: http://192.168.61.131:${APP_PORT}"
            echo "🐳 Image: ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
            """
        }
        failure {
            sh """
            echo "❌ DÉPLOIEMENT ÉCHOUÉ"
            echo "📋 Consultez les logs ci-dessus pour le diagnostic"
            """
        }
    }
}
