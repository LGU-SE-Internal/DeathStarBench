# Attractions Service

The attractions service provides nearby points of interest for hotels, including:
- Restaurants
- Museums
- Cinemas

## API Endpoints

The service exposes three gRPC endpoints through the frontend service:

### 1. Nearby Restaurants
```
GET /restaurants?username=<username>&password=<password>&hotelId=<hotel_id>
```
Returns restaurants near the specified hotel.

### 2. Nearby Museums
```
GET /museums?username=<username>&password=<password>&hotelId=<hotel_id>
```
Returns museums near the specified hotel.

### 3. Nearby Cinemas
```
GET /cinema?username=<username>&password=<password>&hotelId=<hotel_id>
```
Returns cinemas near the specified hotel.

## Database

The attractions service uses MongoDB to store:
- Hotel locations (lat/lon)
- Restaurant data (id, name, location, rating, type)
- Museum data (id, name, location, type)
- Cinema data (id, name, location, type)

Database name: `attractions-db`

Collections:
- `hotels` - Hotel location data
- `restaurants` - Restaurant information
- `museums` - Museum information
- `cinemas` - Cinema information

## Configuration

### Docker Compose
The attractions service is included in `docker-compose.yml`:
- Service port: 8089
- MongoDB port: 27017

### Kubernetes
Deploy the attractions service using:
```bash
kubectl apply -Rf kubernetes/attractions/
```

### Helm
The attractions service is included in the main helm chart as a dependency.

Deploy with:
```bash
helm install hotelreservation helm-chart/hotelreservation/
```

The service will be automatically deployed with:
- Service port: 8089
- MongoDB service port: 27024 (external), 27017 (internal)

## Load Testing

A load testing script with attractions endpoints is available:
```bash
../wrk2/wrk -D exp -t 10 -c 100 -d 30s -L \
  -s ./wrk2/scripts/hotel-reservation/mixed-workload_type_1_with_attractions.lua \
  http://localhost:5000 -R 30
```

The script includes:
- 50% hotel search requests
- 30% recommendation requests
- 6% restaurant lookup requests
- 6% museum lookup requests
- 7% cinema lookup requests
- 0.5% user login requests
- 0.5% reservation requests

## Architecture

The attractions service:
1. Receives requests from the frontend service
2. Looks up hotel location in MongoDB
3. Uses geo-indexing to find nearby attractions within 10km radius
4. Returns up to 5 nearest attractions

The service uses the `go-geoindex` library for efficient geographic queries.
