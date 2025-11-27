import aiohttp
import asyncio
import sys
import json
import argparse

async def upload_cast_info(session, addr, cast):
  try:
    async with session.post(addr + "/wrk2-api/cast-info/write", json=cast, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        print(f"Warning: cast-info write failed with status {resp.status}: {text[:100]}")
      return text
  except Exception as e:
    print(f"Error uploading cast info: {e}")
    return None

async def upload_plot(session, addr, plot):
  try:
    async with session.post(addr + "/wrk2-api/plot/write", json=plot, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        print(f"Warning: plot write failed with status {resp.status}: {text[:100]}")
      return text
  except Exception as e:
    print(f"Error uploading plot: {e}")
    return None

async def upload_movie_info(session, addr, movie):
  try:
    async with session.post(addr + "/wrk2-api/movie-info/write", json=movie, timeout=aiohttp.ClientTimeout(total=30)) as resp:
      text = await resp.text()
      if resp.status != 200:
        print(f"Warning: movie-info write failed with status {resp.status} for movie {movie.get('movie_id')}: {text[:100]}")
      return text
  except Exception as e:
    print(f"Error uploading movie info for {movie.get('movie_id')}: {e}")
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
        print(f"Warning: movie register failed with status {resp.status} for movie {movie.get('movie_id')}: {text[:100]}")
      return text
  except Exception as e:
    print(f"Error registering movie {movie.get('movie_id')}: {e}")
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

async def write_movie_info(addr, raw_movies):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=50)
  async with aiohttp.ClientSession(connector=conn) as session:
    for raw_movie in raw_movies:
      movie = dict()
      casts = list()
      movie["movie_id"] = str(raw_movie["id"])
      movie["title"] = raw_movie["title"]
      movie["plot_id"] = raw_movie["id"]
      for raw_cast in raw_movie["cast"]:
        try:
          cast = dict()
          cast["cast_id"] = raw_cast["cast_id"]
          cast["character"] = raw_cast["character"]
          cast["cast_info_id"] = raw_cast["id"]
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

async def verify_data(addr, movie_ids):
  """Verify that movies were actually written by reading them back."""
  import random
  sample_ids = random.sample(movie_ids, min(5, len(movie_ids)))
  print(f"Verifying data for sample movie IDs: {sample_ids}")
  
  async with aiohttp.ClientSession() as session:
    for movie_id in sample_ids:
      try:
        url = f"{addr}/wrk2-api/movie/read-info?movie_id={movie_id}"
        async with session.get(url, timeout=aiohttp.ClientTimeout(total=10)) as resp:
          text = await resp.text()
          if resp.status == 200 and "title" in text:
            print(f"  ✓ Movie {movie_id} verified successfully")
          else:
            print(f"  ✗ Movie {movie_id} NOT FOUND (status {resp.status}): {text[:100]}")
      except Exception as e:
        print(f"  ✗ Error verifying movie {movie_id}: {e}")

if __name__ == '__main__':
  parser = argparse.ArgumentParser()
  parser.add_argument("-c", "--cast", action="store", dest="cast_filename",
    type=str, default="../datasets/tmdb/casts.json")
  parser.add_argument("-m", "--movie", action="store", dest="movie_filename",
    type=str, default="../datasets/tmdb/movies.json")
  parser.add_argument("--server_address", action="store", dest="server_addr",
    type=str, default="http://127.0.0.1:8080")
  args = parser.parse_args()

  print(f"Server address: {args.server_addr}")
  print(f"Cast file: {args.cast_filename}")
  print(f"Movie file: {args.movie_filename}")

  with open(args.cast_filename, 'r') as cast_file:
    raw_casts = json.load(cast_file)
  print(f"Loaded {len(raw_casts)} casts")
  loop = asyncio.get_event_loop()
  future = asyncio.ensure_future(write_cast_info(args.server_addr, raw_casts))
  loop.run_until_complete(future)

  with open(args.movie_filename, 'r') as movie_file:
    raw_movies = json.load(movie_file)
  print(f"Loaded {len(raw_movies)} movies")
  movie_ids = [str(m["id"]) for m in raw_movies]
  loop = asyncio.get_event_loop()
  future = asyncio.ensure_future(write_movie_info(args.server_addr, raw_movies))
  loop.run_until_complete(future)
  
  # Wait a moment for writes to propagate
  import time
  print("Waiting 5 seconds for writes to propagate...")
  time.sleep(5)
  
  # Verify data was written
  loop = asyncio.get_event_loop()
  future = asyncio.ensure_future(verify_data(args.server_addr, movie_ids))
  loop.run_until_complete(future)