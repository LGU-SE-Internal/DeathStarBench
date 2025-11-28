import aiohttp
import asyncio
import sys
import json
import argparse
import time

# Global counters for tracking success/failure
stats = {
    "cast_info_success": 0,
    "cast_info_failed": 0,
    "movie_info_success": 0,
    "movie_info_failed": 0,
    "plot_success": 0,
    "plot_failed": 0,
    "register_success": 0,
    "register_failed": 0,
}

async def upload_cast_info(session, addr, cast):
  try:
    async with session.post(addr + "/wrk2-api/cast-info/write", json=cast, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        print(f"Warning: cast-info write failed with status {resp.status}: {text[:100]}")
        stats["cast_info_failed"] += 1
        return None
      stats["cast_info_success"] += 1
      return text
  except Exception as e:
    print(f"Error uploading cast info: {e}")
    stats["cast_info_failed"] += 1
    return None

async def upload_plot(session, addr, plot):
  try:
    async with session.post(addr + "/wrk2-api/plot/write", json=plot, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        print(f"Warning: plot write failed with status {resp.status}: {text[:100]}")
        stats["plot_failed"] += 1
        return None
      stats["plot_success"] += 1
      return text
  except Exception as e:
    print(f"Error uploading plot: {e}")
    stats["plot_failed"] += 1
    return None

async def upload_movie_info(session, addr, movie):
  try:
    async with session.post(addr + "/wrk2-api/movie-info/write", json=movie, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        print(f"Warning: movie-info write failed with status {resp.status} for movie {movie.get('movie_id')}: {text[:100]}")
        stats["movie_info_failed"] += 1
        return None
      stats["movie_info_success"] += 1
      return text
  except Exception as e:
    print(f"Error uploading movie info for {movie.get('movie_id')}: {e}")
    stats["movie_info_failed"] += 1
    return None

async def register_movie(session, addr, movie):
  params = {
    "title": movie["title"],
    "movie_id": movie["movie_id"]
  }
  try:
    async with session.post(addr + "/wrk2-api/movie/register", data=params, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        # Check if it's a duplicate title error (which is expected for some movies)
        if "already existed" in text:
          print(f"Info: Movie '{movie.get('title')}' (ID: {movie.get('movie_id')}) already registered (duplicate title in dataset)")
        else:
          print(f"Warning: movie register failed with status {resp.status} for movie {movie.get('movie_id')}: {text[:100]}")
        stats["register_failed"] += 1
        return None
      stats["register_success"] += 1
      return text
  except Exception as e:
    print(f"Error registering movie {movie.get('movie_id')}: {e}")
    stats["register_failed"] += 1
    return None

async def write_cast_info(addr, raw_casts):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=50)
  async with aiohttp.ClientSession(connector=conn) as session:
    for raw_cast in raw_casts:
      try:
        cast = dict()
        cast["cast_info_id"] = raw_cast["id"]
        cast["name"] = raw_cast["name"]
        cast["gender"] = True if raw_cast["gender"] == 2 else False
        cast["intro"] = raw_cast["biography"]
        task = asyncio.ensure_future(upload_cast_info(session, addr, cast))
        tasks.append(task)
        idx += 1
      except:
        print("Warning: cast info missing!")
      if idx % 50 == 0:
        resps = await asyncio.gather(*tasks)
        tasks = []
        print(idx, "casts finished")
    if tasks:
      resps = await asyncio.gather(*tasks)
    print(idx, "casts finished")

async def write_movie_info(addr, raw_movies, valid_cast_ids):
  idx = 0
  tasks = []
  seen_titles = set()  # Track seen titles to handle duplicates
  conn = aiohttp.TCPConnector(limit=50)
  async with aiohttp.ClientSession(connector=conn) as session:
    for raw_movie in raw_movies:
      movie = dict()
      casts = list()
      seen_cast_info_ids = set()  # Track seen cast_info_ids to avoid duplicates
      movie["movie_id"] = str(raw_movie["id"])
      original_title = raw_movie["title"]
      
      # Handle duplicate titles by appending movie_id to make them unique
      # This ensures all movies can be registered in the movie-id collection
      if original_title in seen_titles:
        # Make title unique by appending the movie_id
        movie["title"] = f"{original_title} ({raw_movie['id']})"
        print(f"Info: Duplicate title detected, using unique title: '{movie['title']}' for movie_id {raw_movie['id']}")
      else:
        movie["title"] = original_title
        seen_titles.add(original_title)
      
      movie["plot_id"] = raw_movie["id"]
      for raw_cast in raw_movie["cast"]:
        try:
          cast_info_id = raw_cast["id"]
          # Skip duplicate cast_info_ids to avoid "cast_info_ids are duplicated" error in PageService
          if cast_info_id in seen_cast_info_ids:
            continue
          # Skip cast members that don't exist in casts.json to avoid "return set incomplete" error
          if cast_info_id not in valid_cast_ids:
            continue
          seen_cast_info_ids.add(cast_info_id)
          
          cast = dict()
          cast["cast_id"] = raw_cast["cast_id"]
          cast["character"] = raw_cast["character"]
          cast["cast_info_id"] = cast_info_id
          casts.append(cast)
        except:
          print("Warning: cast info missing!")
      movie["casts"] = casts
      # Filter out None/null values from thumbnail_ids to prevent cjson.null errors in Lua
      movie["thumbnail_ids"] = [raw_movie["poster_path"]] if raw_movie["poster_path"] is not None else []
      movie["photo_ids"] = []
      movie["video_ids"] = []
      movie["avg_rating"] = raw_movie["vote_average"]
      movie["num_rating"] = raw_movie["vote_count"]
      task = asyncio.ensure_future(upload_movie_info(session, addr, movie))
      tasks.append(task)
      plot = dict()
      plot["plot_id"] = raw_movie["id"]
      plot["plot"] = raw_movie["overview"]
      task = asyncio.ensure_future(upload_plot(session, addr, plot))
      tasks.append(task)
      task = asyncio.ensure_future(register_movie(session, addr, movie))
      tasks.append(task)
      idx += 1
      if idx % 50 == 0:
        resps = await asyncio.gather(*tasks)
        tasks = []
        print(idx, "movies finished")
    if tasks:
      resps = await asyncio.gather(*tasks)
    print(idx, "movies finished")

async def verify_data(addr, movie_ids, max_retries=3, retry_delay=5):
  """Verify that movies were actually written by reading them back.
  
  Args:
      addr: Server address
      movie_ids: List of all movie IDs to sample from
      max_retries: Maximum number of retry attempts for failed verifications
      retry_delay: Seconds to wait between retries
  
  Returns:
      tuple: (success_count, failure_count)
  """
  import random
  sample_ids = random.sample(movie_ids, min(5, len(movie_ids)))
  print(f"Verifying data for sample movie IDs: {sample_ids}")
  
  success_count = 0
  failure_count = 0
  
  async with aiohttp.ClientSession() as session:
    for movie_id in sample_ids:
      verified = False
      for attempt in range(max_retries):
        try:
          url = f"{addr}/wrk2-api/movie/read-info?movie_id={movie_id}"
          async with session.get(url, timeout=aiohttp.ClientTimeout(total=10)) as resp:
            text = await resp.text()
            if resp.status == 200 and "title" in text:
              print(f"  ✓ Movie {movie_id} verified successfully")
              verified = True
              success_count += 1
              break
            else:
              if attempt < max_retries - 1:
                print(f"  ⏳ Movie {movie_id} not found (attempt {attempt + 1}/{max_retries}), retrying in {retry_delay}s...")
                await asyncio.sleep(retry_delay)
              else:
                print(f"  ✗ Movie {movie_id} NOT FOUND after {max_retries} attempts (status {resp.status}): {text[:100]}")
        except Exception as e:
          if attempt < max_retries - 1:
            print(f"  ⏳ Error verifying movie {movie_id} (attempt {attempt + 1}/{max_retries}): {e}, retrying...")
            await asyncio.sleep(retry_delay)
          else:
            print(f"  ✗ Error verifying movie {movie_id} after {max_retries} attempts: {e}")
      
      if not verified:
        failure_count += 1
  
  return success_count, failure_count

async def wait_for_services(addr, services_to_check, timeout=60):
  """Wait for backend services to be ready by checking health endpoints.
  
  Args:
      addr: Server address (nginx)
      services_to_check: List of service endpoints to check
      timeout: Maximum seconds to wait
  """
  print(f"Waiting for backend services to be ready (timeout: {timeout}s)...")
  
  start_time = time.time()
  async with aiohttp.ClientSession() as session:
    while time.time() - start_time < timeout:
      all_ready = True
      for endpoint in services_to_check:
        try:
          url = f"{addr}{endpoint}"
          # Send a test request to check if the endpoint is responsive
          async with session.get(url, timeout=aiohttp.ClientTimeout(total=5)) as resp:
            # Any response (even 400/404) means the service is running
            # We just want to make sure we can connect
            if resp.status >= 500:
              all_ready = False
              break
        except Exception as e:
          print(f"  Service at {endpoint} not ready: {e}")
          all_ready = False
          break
      
      if all_ready:
        print("  All services are ready!")
        return True
      
      print(f"  Waiting for services... ({int(time.time() - start_time)}s elapsed)")
      await asyncio.sleep(5)
  
  print(f"  Warning: Timeout waiting for services after {timeout}s, proceeding anyway...")
  return False

if __name__ == '__main__':
  parser = argparse.ArgumentParser()
  parser.add_argument("-c", "--cast", action="store", dest="cast_filename",
    type=str, default="../datasets/tmdb/casts.json")
  parser.add_argument("-m", "--movie", action="store", dest="movie_filename",
    type=str, default="../datasets/tmdb/movies.json")
  parser.add_argument("--server_address", action="store", dest="server_addr",
    type=str, default="http://127.0.0.1:8080")
  parser.add_argument("--wait-timeout", action="store", dest="wait_timeout",
    type=int, default=60, help="Timeout in seconds to wait for services to be ready")
  args = parser.parse_args()

  print(f"Server address: {args.server_addr}")
  print(f"Cast file: {args.cast_filename}")
  print(f"Movie file: {args.movie_filename}")

  # Wait for backend services to be ready before starting data initialization
  loop = asyncio.get_event_loop()
  # Check that we can reach the nginx endpoints (which proxy to backend services)
  services_to_check = [
    "/wrk2-api/movie/read-info?movie_id=test",  # This will return 404/500 but proves connectivity
  ]
  future = asyncio.ensure_future(wait_for_services(args.server_addr, services_to_check, args.wait_timeout))
  loop.run_until_complete(future)

  with open(args.cast_filename, 'r') as cast_file:
    raw_casts = json.load(cast_file)
  print(f"Loaded {len(raw_casts)} casts")
  # Build set of valid cast IDs to filter movie casts
  valid_cast_ids = set(c['id'] for c in raw_casts)
  future = asyncio.ensure_future(write_cast_info(args.server_addr, raw_casts))
  loop.run_until_complete(future)

  with open(args.movie_filename, 'r') as movie_file:
    raw_movies = json.load(movie_file)
  print(f"Loaded {len(raw_movies)} movies")
  movie_ids = [str(m["id"]) for m in raw_movies]
  future = asyncio.ensure_future(write_movie_info(args.server_addr, raw_movies, valid_cast_ids))
  loop.run_until_complete(future)
  
  # Print statistics
  print("\n" + "="*50)
  print("Data Initialization Statistics:")
  print("="*50)
  print(f"  Cast Info:   {stats['cast_info_success']} success, {stats['cast_info_failed']} failed")
  print(f"  Movie Info:  {stats['movie_info_success']} success, {stats['movie_info_failed']} failed")
  print(f"  Plot:        {stats['plot_success']} success, {stats['plot_failed']} failed")
  print(f"  Register:    {stats['register_success']} success, {stats['register_failed']} failed")
  print("="*50)
  
  # Wait for writes to propagate using async sleep
  async def wait_for_propagation():
    print("\nWaiting 10 seconds for writes to propagate...")
    await asyncio.sleep(10)
  
  loop.run_until_complete(wait_for_propagation())
  
  # Verify data was written with retry logic
  print("\nVerifying data was written correctly...")
  future = asyncio.ensure_future(verify_data(args.server_addr, movie_ids, max_retries=3, retry_delay=5))
  success, failure = loop.run_until_complete(future)
  
  print("\n" + "="*50)
  print("Verification Summary:")
  print("="*50)
  print(f"  Verified: {success}/5 movies found")
  if failure > 0:
    print(f"  Warning: {failure} movies could not be verified")
    print("  This may indicate an issue with data initialization or service connectivity")
  else:
    print("  All sample movies verified successfully!")
  print("="*50)
  
  # Exit with error if verification failed
  if failure > 0:
    sys.exit(1)