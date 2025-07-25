import asyncio


async def run_worker():
    from unionai_actor_executor._lib import executor
    await executor()


async def rust_sleeper(seconds: int):
    print(f"Sleeping for {seconds} seconds...")
    from unionai_actor_executor._lib import rust_sleep
    await rust_sleep()


async def rust_test():
    print("[Actor Core] Running Actor tester...")
    from unionai_actor_executor._lib import tester
    _ = await tester()
    print("[Actor core] Actor tester completed")


def main():
    asyncio.run(run_worker())


def tester_main():
    asyncio.run(rust_test())


def sleeper_main():
    asyncio.run(rust_sleeper(2))


if __name__ == "__main__":
    main()
