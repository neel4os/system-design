# Agent, Workflow and Mayhem

## Table of Contents
- [Agent, Workflow and Mayhem](#agent-workflow-and-mayhem)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Building Agent: Nightmare dressed as Daydream](#building-agent-nightmare-dressed-as-daydream)
  - [Graph Execution](#graph-execution)
    - [Philosophies of Graph Exeuction Model](#philosophies-of-graph-exeuction-model)
  - [Pregel](#pregel)
- [References](#references)

## Introduction

Although Agentic AI is redefining software development, yet its difficult to find a definition of agent. A general consesus converged that agent are autonomous system that independently accomplish tasks. According to Anthropic, an agent can be desribed like

<img src=./figure1.svg width=400>

A useful starting point for understanding an agent is an LLM equipped with tools and a loop that allows the LLM to decide what to do next. For instance we call an LLM, and it inturn call the tool defined as and when it is needed. Tool can be considered as environment which returns some information passed back to LLM. The loop continues until a pre defined limit reached or LLM decides no more tool calling needed and we have received the output.

## Building Agent: Nightmare dressed as Daydream

Building agent is not difficult. Any LLM have straight forward API defined. We can define tool and a simple loop with a maximum iteration value with tool calling will implement this. 

```
def my_awesome_agent(user_input: str):
    counter = 0
    while counter < MAX_COUNTER:
        response = call_llm(user_input)
        if response.has_tool_call():
            result = call_tool(response.tool_call)
            user_input = result
            counter += 1
        else:
            return response.answer

    return "Maximum iterations reached"
```
and thats our agent.

Then why is this blog? Why there are new Agent Framework coming out everyday? Because making the loop useful in production is where the mayhem begins..

- Workflow: 

    Real-world tasks rarely consist of a single agent decision. They usually involve multiple steps, branching paths, loops, and sometimes parallel execution. A workflow provides the structure for orchestrating these steps toward a goal.
    For example following is a CVE Fix agent

    <img src=./figure2.svg height=400>

- Crash Resistance:

    What if an agent creashes during the exection? Should we start from scratch again or should be build persistence around the agent?

- Human intervention

    What if agent needs humand approval? Where does the waiting state live?

- What if there are multiple agent

   Some problems benefit from multiple specialized agents. Now we need coordination, communication, task assignment, and shared or isolated state.

And these aren't the only problems. We eventually need observability, evaluation, retries, state management, memory, tool management, concurrency, and a way to resume execution after an interruption. So as we can see the tiny loop is slowly turning into a runtime and this is where agent frameworks enter the picture.

They provide the machinery around the agent: executing steps, managing state, coordinating work, handling interruptions, and providing the infrastructure required to run agents reliably.

There are no shortage of Agentic AI Framework in the current situation. But objective of this write up is not to understand their differences or their use cases. The Objective is to take a deep dive into the machinery that powers this frameworks and to do that, we are going to write our own tiny agentic framework. We will name this project 
```
Agentiq
```

## Graph Execution

Almost core logic of all agentic framework runs on a graph engine because in general, workflows represent a graph. Each node in the graph either can be an agent or it can be a deterministic business logic tied to the domain. Later we will discover [an workflow can be used as an agent](https://learn.microsoft.com/en-us/agent-framework/workflows/as-agents?pivots=programming-language-python)

The graph engine is indepepndent of agent. It has no idea on what the agent is, what tool it call, what LLM it uses. Cosnider the graph engine as mechanism and it does not break when we bring a new agent or a new LLM. Agents are just one or more nodes in the workflow.

So `agentiq` will first focus on building a graph engine. 

### Philosophies of Graph Exeuction Model

Before we write down the graph engine for the `agentiq`, let us try to contemplate on what is the basic difference from a standard graph engine we need to implement.

In a normal graph engine, we focus on `traversing` which is looping over the nodes and try to achieve the objective that was predefined. A traversal is basically a graph, frontier(current node) and visited set. This causes several issues with agentic graph execution. In an agentic graph execution, we need 
- parallezation (calling multiple nodes in the same time), 
- state snapshot (so that in case any nodes crashed, we do not need to start from scratch)

To make this possible, we look into something thats is a different type of graph execution. Instead of traversing, we should focus on converging. What converging actually does here is trying to achieve a state that stopped changing. Consider a spreadsheet with lots of formula. When you chnage any number, every cell recomputes and then since the state of spreadsheet does not change anymore, it achieved all mutation that is needed. It does not traverse.

so putting the comparison of basic graph execution between these two philosophies are the following

|  | Traversal | Convergence |
| --- | --- | --- |
| Question asked | "where do I go next" | "did anything changed"|
| Frontier | A plan | A record |
| Direction | push - A says run B | pull - B says "I watch A" |
| Termination | structural : ran out of queues of nodes | semantic : reached a fix point |
| Re Execution | hazardous | the core mechanism |

If it is not be clear this different philosphies of graph execution, this mechanism of converge is published by google in their `pregel` 
This is not a tutorial of pregel but we want to assert that pregel's vertex centric graph processing model is extremely apt for agentic graph execution. In next section we will be focusing on very short introduction of `pregel` and how it can be used to create a graph engine for `agentiq`. 

## Pregel

Pregel is model for computing over very large graphs. Its central idea is 
> Think like a vertex

So instead of writing a program that loops over the graph, you write a function describing what one vertex does when it wakes up. The runtime runs the function for every vertex, in rounds

so the graph computationis is a sequence of `superstep`. each active vertex 

1. Runs `compute(message)` - in parallel, independently of every other vertex




# References
- [langgraph blog on How to think about agent frameworks](https://www.langchain.com/blog/how-to-think-about-agent-frameworks)
  
- [How to build effective agent by Anthropic](https://www.anthropic.com/engineering/building-effective-agents)

- [Pregel origial paper](https://15799.courses.cs.cmu.edu/fall2013/static/papers/p135-malewicz.pdf)







